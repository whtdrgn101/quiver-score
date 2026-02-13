import base64

import httpx
from fastapi import APIRouter, Depends, HTTPException, UploadFile
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.club import Club, ClubMember, ClubTeam, ClubTeamMember
from app.models.scoring import ScoringSession
from app.schemas.user import AvatarUrlUpload, ProfileClubOut, ProfileClubTeamOut, UserOut, UserUpdate, PublicProfileOut, PublicSessionSummary
from app.core.exceptions import NotFoundError

router = APIRouter(prefix="/users", tags=["users"])

ALLOWED_TYPES = {"image/jpeg", "image/png", "image/webp"}
MAX_AVATAR_BYTES = 2 * 1024 * 1024  # 2 MB


@router.get("/me", response_model=UserOut)
async def get_me(user: User = Depends(get_current_user)):
    return user


@router.patch("/me", response_model=UserOut)
async def update_me(
    body: UserUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    for field, value in body.model_dump(exclude_unset=True).items():
        setattr(user, field, value)
    await db.commit()
    await db.refresh(user)
    return user


@router.post("/me/avatar", response_model=UserOut)
async def upload_avatar(
    file: UploadFile,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    if file.content_type not in ALLOWED_TYPES:
        raise HTTPException(status_code=400, detail="File must be JPEG, PNG, or WebP")
    data = await file.read()
    if len(data) > MAX_AVATAR_BYTES:
        raise HTTPException(status_code=400, detail="File must be under 2 MB")
    encoded = base64.b64encode(data).decode()
    user.avatar = f"data:{file.content_type};base64,{encoded}"
    await db.commit()
    await db.refresh(user)
    return user


@router.post("/me/avatar-url", response_model=UserOut)
async def upload_avatar_from_url(
    body: AvatarUrlUpload,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    try:
        async with httpx.AsyncClient(follow_redirects=True, timeout=10) as client:
            resp = await client.get(str(body.url))
            resp.raise_for_status()
    except httpx.HTTPError:
        raise HTTPException(status_code=400, detail="Could not fetch image from URL")

    content_type = resp.headers.get("content-type", "").split(";")[0].strip()
    if content_type not in ALLOWED_TYPES:
        raise HTTPException(status_code=400, detail="URL must point to a JPEG, PNG, or WebP image")
    if len(resp.content) > MAX_AVATAR_BYTES:
        raise HTTPException(status_code=400, detail="Image must be under 2 MB")

    encoded = base64.b64encode(resp.content).decode()
    user.avatar = f"data:{content_type};base64,{encoded}"
    await db.commit()
    await db.refresh(user)
    return user


@router.delete("/me/avatar", response_model=UserOut)
async def delete_avatar(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    user.avatar = None
    await db.commit()
    await db.refresh(user)
    return user


@router.get("/me/clubs", response_model=list[ProfileClubOut])
async def get_my_clubs_with_teams(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    memberships_result = await db.execute(
        select(ClubMember).where(ClubMember.user_id == user.id)
    )
    memberships = memberships_result.scalars().all()
    clubs_out = []
    for membership in memberships:
        club_result = await db.execute(select(Club).where(Club.id == membership.club_id))
        club = club_result.scalar_one_or_none()
        if not club:
            continue
        team_memberships = await db.execute(
            select(ClubTeamMember).where(ClubTeamMember.user_id == user.id)
        )
        team_members = team_memberships.scalars().all()
        teams_out = []
        for tm in team_members:
            team_result = await db.execute(
                select(ClubTeam).where(ClubTeam.id == tm.team_id, ClubTeam.club_id == club.id)
            )
            team = team_result.scalar_one_or_none()
            if team:
                teams_out.append(ProfileClubTeamOut(team_id=team.id, team_name=team.name))
        clubs_out.append(ProfileClubOut(
            club_id=club.id, club_name=club.name, role=membership.role, teams=teams_out,
        ))
    return clubs_out


@router.get("/{username}", response_model=PublicProfileOut)
async def get_public_profile(
    username: str,
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(select(User).where(User.username == username))
    user = result.scalar_one_or_none()
    if not user or not user.profile_public:
        raise NotFoundError("Profile not found")

    # Compute stats from completed sessions
    completed = [s for s in user.scoring_sessions if s.status == "completed"]
    total_sessions = len(user.scoring_sessions)
    completed_sessions = len(completed)
    total_arrows = sum(s.total_arrows for s in completed)
    total_x_count = sum(s.total_x_count for s in completed)

    personal_best_score = None
    personal_best_template = None
    if completed:
        best = max(completed, key=lambda s: s.total_score)
        personal_best_score = best.total_score
        personal_best_template = best.template.name if best.template else None

    # Recent 5 completed sessions
    recent = sorted(completed, key=lambda s: s.completed_at or s.started_at, reverse=True)[:5]
    recent_sessions = [
        PublicSessionSummary(
            template_name=s.template.name if s.template else None,
            total_score=s.total_score,
            total_x_count=s.total_x_count,
            total_arrows=s.total_arrows,
            completed_at=s.completed_at,
            share_token=s.share_token,
        )
        for s in recent
    ]

    # Clubs and teams
    memberships_result = await db.execute(
        select(ClubMember).where(ClubMember.user_id == user.id)
    )
    memberships = memberships_result.scalars().all()
    clubs_out = []
    for membership in memberships:
        club_result = await db.execute(select(Club).where(Club.id == membership.club_id))
        club = club_result.scalar_one_or_none()
        if not club:
            continue
        # Find teams in this club the user belongs to
        team_memberships = await db.execute(
            select(ClubTeamMember).where(ClubTeamMember.user_id == user.id)
        )
        team_members = team_memberships.scalars().all()
        teams_out = []
        for tm in team_members:
            team_result = await db.execute(
                select(ClubTeam).where(ClubTeam.id == tm.team_id, ClubTeam.club_id == club.id)
            )
            team = team_result.scalar_one_or_none()
            if team:
                teams_out.append(ProfileClubTeamOut(team_id=team.id, team_name=team.name))
        clubs_out.append(ProfileClubOut(
            club_id=club.id, club_name=club.name, role=membership.role, teams=teams_out,
        ))

    return PublicProfileOut(
        id=user.id,
        username=user.username,
        display_name=user.display_name,
        bow_type=user.bow_type,
        bio=user.bio,
        avatar=user.avatar,
        created_at=user.created_at,
        total_sessions=total_sessions,
        completed_sessions=completed_sessions,
        total_arrows=total_arrows,
        total_x_count=total_x_count,
        personal_best_score=personal_best_score,
        personal_best_template=personal_best_template,
        recent_sessions=recent_sessions,
        clubs=clubs_out,
    )
