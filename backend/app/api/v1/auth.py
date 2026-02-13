import uuid

from fastapi import APIRouter, Depends, Request, status
from sqlalchemy import delete, select, or_
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.rate_limit import limiter

from app.database import get_db
from app.models.user import User
from app.models.scoring import ScoringSession, End, Arrow, PersonalRecord
from app.models.equipment import Equipment
from app.models.setup_profile import SetupProfile, SetupEquipment
from app.models.round_template import RoundTemplate, RoundTemplateStage
from app.models.club import (
    Club, ClubMember, ClubInvite, ClubEvent, ClubEventParticipant,
    ClubTeam, ClubTeamMember, ClubSharedRound,
)
from app.models.notification import Notification
from app.models.classification import ClassificationRecord
from app.models.sight_mark import SightMark
from app.models.tournament import Tournament, TournamentParticipant
from app.models.coaching import CoachAthleteLink, SessionAnnotation
from app.models.social import Follow, FeedItem
from app.dependencies import get_current_user
from pydantic import BaseModel, Field
from app.schemas.auth import (
    RegisterRequest, LoginRequest, TokenResponse, RefreshRequest, PasswordChange,
    ForgotPasswordRequest, ResetPasswordRequest, VerifyEmailRequest, MessageResponse,
)
from app.core.security import (
    hash_password, verify_password, create_access_token, create_refresh_token,
    decode_token, create_reset_token, verify_reset_token,
    create_email_verification_token, verify_email_verification_token,
)
from app.core.email import send_password_reset_email, send_verification_email
from app.core.exceptions import AuthError, ConflictError

router = APIRouter(prefix="/auth", tags=["auth"])


@router.post("/register", response_model=TokenResponse, status_code=status.HTTP_201_CREATED)
@limiter.limit("5/minute")
async def register(request: Request, body: RegisterRequest, db: AsyncSession = Depends(get_db)):
    result = await db.execute(
        select(User).where(or_(User.email == body.email, User.username == body.username))
    )
    if result.scalar_one_or_none():
        raise ConflictError("Email or username already registered")

    verification_token = create_email_verification_token(body.email)

    user = User(
        email=body.email,
        username=body.username,
        hashed_password=hash_password(body.password),
        display_name=body.display_name or body.username,
        email_verification_token=verification_token,
    )
    db.add(user)
    await db.commit()
    await db.refresh(user)

    send_verification_email(user.email, verification_token)

    return TokenResponse(
        access_token=create_access_token(str(user.id)),
        refresh_token=create_refresh_token(str(user.id)),
    )


@router.post("/login", response_model=TokenResponse)
@limiter.limit("5/minute")
async def login(request: Request, body: LoginRequest, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(User).where(User.username == body.username))
    user = result.scalar_one_or_none()
    if not user or not verify_password(body.password, user.hashed_password):
        raise AuthError("Invalid username or password")

    return TokenResponse(
        access_token=create_access_token(str(user.id)),
        refresh_token=create_refresh_token(str(user.id)),
    )


@router.post("/refresh", response_model=TokenResponse)
@limiter.limit("10/minute")
async def refresh(request: Request, body: RefreshRequest, db: AsyncSession = Depends(get_db)):
    payload = decode_token(body.refresh_token)
    if payload is None or payload.get("type") != "refresh":
        raise AuthError("Invalid refresh token")

    try:
        user_id = uuid.UUID(payload["sub"])
    except (KeyError, ValueError):
        raise AuthError("Invalid refresh token")

    result = await db.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise AuthError("User not found")

    return TokenResponse(
        access_token=create_access_token(str(user.id)),
        refresh_token=create_refresh_token(str(user.id)),
    )


@router.post("/verify-email", response_model=MessageResponse)
@limiter.limit("5/minute")
async def verify_email(request: Request, body: VerifyEmailRequest, db: AsyncSession = Depends(get_db)):
    email = verify_email_verification_token(body.token)
    if not email:
        raise AuthError("Invalid or expired verification token")

    result = await db.execute(select(User).where(User.email == email))
    user = result.scalar_one_or_none()
    if not user:
        raise AuthError("Invalid or expired verification token")

    user.email_verified = True
    user.email_verification_token = None
    await db.commit()
    return MessageResponse(detail="Email verified successfully.")


@router.post("/resend-verification", response_model=MessageResponse)
async def resend_verification(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    if user.email_verified:
        return MessageResponse(detail="Email is already verified.")

    token = create_email_verification_token(user.email)
    user.email_verification_token = token
    await db.commit()
    send_verification_email(user.email, token)
    return MessageResponse(detail="Verification email sent.")


@router.post("/change-password", status_code=status.HTTP_200_OK)
async def change_password(
    body: PasswordChange,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    if not verify_password(body.current_password, user.hashed_password):
        raise AuthError("Current password is incorrect")

    user.hashed_password = hash_password(body.new_password)
    await db.commit()
    return {"detail": "Password changed successfully"}


@router.post("/forgot-password", response_model=MessageResponse)
@limiter.limit("5/minute")
async def forgot_password(request: Request, body: ForgotPasswordRequest, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(User).where(User.email == body.email))
    user = result.scalar_one_or_none()
    if user:
        token = create_reset_token(user.email)
        send_password_reset_email(user.email, token)
    return MessageResponse(detail="If that email is registered, you'll receive a reset link shortly.")


@router.post("/reset-password", response_model=MessageResponse)
@limiter.limit("3/minute")
async def reset_password(request: Request, body: ResetPasswordRequest, db: AsyncSession = Depends(get_db)):
    email = verify_reset_token(body.token)
    if not email:
        raise AuthError("Invalid or expired reset token")
    result = await db.execute(select(User).where(User.email == email))
    user = result.scalar_one_or_none()
    if not user:
        raise AuthError("Invalid or expired reset token")
    user.hashed_password = hash_password(body.new_password)
    await db.commit()
    return MessageResponse(detail="Password reset successfully. You can now sign in.")


class DeleteAccountRequest(BaseModel):
    confirmation: str = Field(..., min_length=1)


@router.post("/delete-account", status_code=status.HTTP_200_OK)
async def delete_account(
    body: DeleteAccountRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    if body.confirmation != "Yes, delete ALL of my data":
        raise AuthError("Confirmation text does not match")

    uid = user.id

    # 1. Scoring data: arrows → ends → sessions, personal records
    session_ids_result = await db.execute(
        select(ScoringSession.id).where(ScoringSession.user_id == uid)
    )
    session_ids = list(session_ids_result.scalars().all())

    if session_ids:
        end_ids_result = await db.execute(
            select(End.id).where(End.session_id.in_(session_ids))
        )
        end_ids = list(end_ids_result.scalars().all())
        if end_ids:
            await db.execute(delete(Arrow).where(Arrow.end_id.in_(end_ids)))
        await db.execute(delete(End).where(End.session_id.in_(session_ids)))

    # Session annotations (authored by user or on user's sessions)
    conditions = [SessionAnnotation.author_id == uid]
    if session_ids:
        conditions.append(SessionAnnotation.session_id.in_(session_ids))
    await db.execute(delete(SessionAnnotation).where(or_(*conditions)))

    # Tournament participants referencing user's sessions
    await db.execute(delete(TournamentParticipant).where(TournamentParticipant.user_id == uid))

    await db.execute(delete(PersonalRecord).where(PersonalRecord.user_id == uid))
    await db.execute(delete(ScoringSession).where(ScoringSession.user_id == uid))

    # 2. Equipment and setups
    setup_ids_result = await db.execute(
        select(SetupProfile.id).where(SetupProfile.user_id == uid)
    )
    setup_ids = list(setup_ids_result.scalars().all())
    if setup_ids:
        await db.execute(delete(SetupEquipment).where(SetupEquipment.setup_id.in_(setup_ids)))
    await db.execute(delete(SetupProfile).where(SetupProfile.user_id == uid))
    await db.execute(delete(Equipment).where(Equipment.user_id == uid))

    # 3. Coaching
    await db.execute(delete(CoachAthleteLink).where(
        or_(CoachAthleteLink.coach_id == uid, CoachAthleteLink.athlete_id == uid)
    ))

    # 4. Clubs owned by user — delete entire club and all sub-entities
    owned_club_ids_result = await db.execute(
        select(Club.id).where(Club.owner_id == uid)
    )
    owned_club_ids = list(owned_club_ids_result.scalars().all())

    if owned_club_ids:
        # Tournament participants for club tournaments
        tournament_ids_result = await db.execute(
            select(Tournament.id).where(Tournament.club_id.in_(owned_club_ids))
        )
        tournament_ids = list(tournament_ids_result.scalars().all())
        if tournament_ids:
            await db.execute(delete(TournamentParticipant).where(
                TournamentParticipant.tournament_id.in_(tournament_ids)
            ))
        await db.execute(delete(Tournament).where(Tournament.club_id.in_(owned_club_ids)))

        # Club events and participants
        event_ids_result = await db.execute(
            select(ClubEvent.id).where(ClubEvent.club_id.in_(owned_club_ids))
        )
        event_ids = list(event_ids_result.scalars().all())
        if event_ids:
            await db.execute(delete(ClubEventParticipant).where(
                ClubEventParticipant.event_id.in_(event_ids)
            ))
        await db.execute(delete(ClubEvent).where(ClubEvent.club_id.in_(owned_club_ids)))

        # Teams and team members
        team_ids_result = await db.execute(
            select(ClubTeam.id).where(ClubTeam.club_id.in_(owned_club_ids))
        )
        team_ids = list(team_ids_result.scalars().all())
        if team_ids:
            await db.execute(delete(ClubTeamMember).where(ClubTeamMember.team_id.in_(team_ids)))
        await db.execute(delete(ClubTeam).where(ClubTeam.club_id.in_(owned_club_ids)))

        await db.execute(delete(ClubSharedRound).where(ClubSharedRound.club_id.in_(owned_club_ids)))
        await db.execute(delete(ClubInvite).where(ClubInvite.club_id.in_(owned_club_ids)))
        await db.execute(delete(ClubMember).where(ClubMember.club_id.in_(owned_club_ids)))
        await db.execute(delete(Club).where(Club.id.in_(owned_club_ids)))

    # 5. Tournaments organized by user (not in owned clubs)
    user_tournament_ids_result = await db.execute(
        select(Tournament.id).where(Tournament.organizer_id == uid)
    )
    user_tournament_ids = list(user_tournament_ids_result.scalars().all())
    if user_tournament_ids:
        await db.execute(delete(TournamentParticipant).where(
            TournamentParticipant.tournament_id.in_(user_tournament_ids)
        ))
        await db.execute(delete(Tournament).where(Tournament.id.in_(user_tournament_ids)))

    # 6. Club participation (non-owned clubs)
    await db.execute(delete(ClubEventParticipant).where(ClubEventParticipant.user_id == uid))
    await db.execute(delete(ClubTeamMember).where(ClubTeamMember.user_id == uid))
    await db.execute(delete(ClubSharedRound).where(ClubSharedRound.shared_by == uid))
    await db.execute(delete(ClubInvite).where(ClubInvite.created_by == uid))
    await db.execute(delete(ClubMember).where(ClubMember.user_id == uid))

    # 7. Custom round templates and stages
    template_ids_result = await db.execute(
        select(RoundTemplate.id).where(
            RoundTemplate.created_by == uid, RoundTemplate.is_official == False
        )
    )
    template_ids = list(template_ids_result.scalars().all())
    if template_ids:
        await db.execute(delete(ClubSharedRound).where(ClubSharedRound.template_id.in_(template_ids)))
        await db.execute(delete(RoundTemplateStage).where(RoundTemplateStage.template_id.in_(template_ids)))
        await db.execute(delete(RoundTemplate).where(RoundTemplate.id.in_(template_ids)))

    # 8. Remaining user data (some have CASCADE but explicit for SQLite)
    await db.execute(delete(Notification).where(Notification.user_id == uid))
    await db.execute(delete(ClassificationRecord).where(ClassificationRecord.user_id == uid))
    await db.execute(delete(SightMark).where(SightMark.user_id == uid))
    await db.execute(delete(Follow).where(or_(Follow.follower_id == uid, Follow.following_id == uid)))
    await db.execute(delete(FeedItem).where(FeedItem.user_id == uid))

    # 9. Delete the user
    await db.execute(delete(User).where(User.id == uid))
    await db.commit()

    return {"detail": "Account and all data deleted"}
