import secrets
import uuid
from datetime import datetime, timedelta, timezone

from fastapi import APIRouter, Depends, Query, Request, status
from sqlalchemy import and_, func, select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.dependencies import get_current_user
from app.core.exceptions import AuthError, ConflictError, NotFoundError, ValidationError
from app.models.club import Club, ClubEvent, ClubEventParticipant, ClubInvite, ClubMember, ClubTeam, ClubTeamMember
from app.models.scoring import PersonalRecord, ScoringSession
from app.models.user import User
from app.schemas.club import (
    ActivityItem,
    ClubCreate,
    ClubDetailOut,
    ClubMemberOut,
    ClubOut,
    ClubUpdate,
    EventCreate,
    EventOut,
    EventParticipantOut,
    EventRSVP,
    EventUpdate,
    InviteCreate,
    InviteOut,
    JoinResult,
    LeaderboardEntry,
    LeaderboardOut,
    TeamCreate,
    TeamDetailOut,
    TeamMemberOut,
    TeamOut,
    TeamUpdate,
)

router = APIRouter(prefix="/clubs", tags=["clubs"])


# ── Helpers ───────────────────────────────────────────────────────────


async def _get_club_member(db: AsyncSession, club_id: uuid.UUID, user_id: uuid.UUID) -> ClubMember:
    result = await db.execute(
        select(ClubMember).where(ClubMember.club_id == club_id, ClubMember.user_id == user_id)
    )
    member = result.scalar_one_or_none()
    if not member:
        # Check if club exists at all
        club_exists = await db.execute(select(Club.id).where(Club.id == club_id))
        if not club_exists.scalar_one_or_none():
            raise NotFoundError("Club not found")
        raise AuthError("You are not a member of this club")
    return member


def _club_out(club: Club, my_role: str | None = None) -> ClubOut:
    return ClubOut(
        id=club.id,
        name=club.name,
        description=club.description,
        avatar=club.avatar,
        owner_id=club.owner_id,
        member_count=len(club.members),
        my_role=my_role,
        created_at=club.created_at,
    )


def _club_detail_out(club: Club, my_role: str | None = None) -> ClubDetailOut:
    members = [
        ClubMemberOut(
            user_id=m.user_id,
            username=m.user.username,
            display_name=m.user.display_name,
            avatar=m.user.avatar,
            role=m.role,
            joined_at=m.joined_at,
        )
        for m in club.members
    ]
    return ClubDetailOut(
        id=club.id,
        name=club.name,
        description=club.description,
        avatar=club.avatar,
        owner_id=club.owner_id,
        member_count=len(club.members),
        my_role=my_role,
        created_at=club.created_at,
        members=members,
    )


def _invite_out(invite: ClubInvite, base_url: str) -> InviteOut:
    return InviteOut(
        id=invite.id,
        code=invite.code,
        url=f"{base_url}/clubs/join/{invite.code}",
        max_uses=invite.max_uses,
        use_count=invite.use_count,
        expires_at=invite.expires_at,
        active=invite.active,
        created_at=invite.created_at,
    )


async def _get_member_user_ids(db: AsyncSession, club_id: uuid.UUID) -> list[uuid.UUID]:
    result = await db.execute(select(ClubMember.user_id).where(ClubMember.club_id == club_id))
    return list(result.scalars().all())


# ── Club CRUD ─────────────────────────────────────────────────────────


@router.post("", response_model=ClubOut, status_code=status.HTTP_201_CREATED)
async def create_club(
    body: ClubCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    club = Club(name=body.name, description=body.description, owner_id=user.id)
    db.add(club)
    await db.flush()

    member = ClubMember(club_id=club.id, user_id=user.id, role="owner")
    db.add(member)
    await db.commit()
    await db.refresh(club)
    return _club_out(club, my_role="owner")


@router.get("", response_model=list[ClubOut])
async def list_my_clubs(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ClubMember).where(ClubMember.user_id == user.id)
    )
    memberships = result.scalars().all()
    clubs = []
    for m in memberships:
        result = await db.execute(select(Club).where(Club.id == m.club_id))
        club = result.scalar_one_or_none()
        if club:
            clubs.append(_club_out(club, my_role=m.role))
    return clubs


@router.get("/join/{code}", response_model=ClubOut)
async def preview_invite(
    code: str,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    invite = await _get_valid_invite(db, code)
    result = await db.execute(select(Club).where(Club.id == invite.club_id))
    club = result.scalar_one_or_none()
    if not club:
        raise NotFoundError("Club not found")
    # Check if already a member
    existing = await db.execute(
        select(ClubMember).where(ClubMember.club_id == club.id, ClubMember.user_id == user.id)
    )
    my_role = None
    em = existing.scalar_one_or_none()
    if em:
        my_role = em.role
    return _club_out(club, my_role=my_role)


@router.post("/join/{code}", response_model=JoinResult)
async def join_club(
    code: str,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    invite = await _get_valid_invite(db, code)

    # Check already a member
    existing = await db.execute(
        select(ClubMember).where(ClubMember.club_id == invite.club_id, ClubMember.user_id == user.id)
    )
    if existing.scalar_one_or_none():
        raise ConflictError("You are already a member of this club")

    member = ClubMember(club_id=invite.club_id, user_id=user.id, role="member")
    db.add(member)
    invite.use_count += 1
    await db.commit()

    result = await db.execute(select(Club).where(Club.id == invite.club_id))
    club = result.scalar_one()
    return JoinResult(club_id=club.id, club_name=club.name, role="member")


async def _get_valid_invite(db: AsyncSession, code: str) -> ClubInvite:
    result = await db.execute(select(ClubInvite).where(ClubInvite.code == code, ClubInvite.active == True))
    invite = result.scalar_one_or_none()
    if not invite:
        raise NotFoundError("Invite not found or expired")
    now = datetime.now(timezone.utc)
    if invite.expires_at and invite.expires_at < now:
        raise NotFoundError("Invite has expired")
    if invite.max_uses and invite.use_count >= invite.max_uses:
        raise NotFoundError("Invite has reached maximum uses")
    return invite


@router.get("/{club_id}", response_model=ClubDetailOut)
async def get_club(
    club_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    result = await db.execute(select(Club).where(Club.id == club_id))
    club = result.scalar_one()
    return _club_detail_out(club, my_role=member.role)


@router.patch("/{club_id}", response_model=ClubOut)
async def update_club(
    club_id: uuid.UUID,
    body: ClubUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role != "owner":
        raise AuthError("Only the club owner can update the club")
    result = await db.execute(select(Club).where(Club.id == club_id))
    club = result.scalar_one()
    if body.name is not None:
        club.name = body.name
    if body.description is not None:
        club.description = body.description
    await db.commit()
    await db.refresh(club)
    return _club_out(club, my_role=member.role)


@router.delete("/{club_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_club(
    club_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role != "owner":
        raise AuthError("Only the club owner can delete the club")
    result = await db.execute(select(Club).where(Club.id == club_id))
    club = result.scalar_one()
    await db.delete(club)
    await db.commit()


# ── Invites ───────────────────────────────────────────────────────────


@router.post("/{club_id}/invites", response_model=InviteOut, status_code=status.HTTP_201_CREATED)
async def create_invite(
    club_id: uuid.UUID,
    body: InviteCreate,
    request: Request,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can create invites")

    expires_at = None
    if body.expires_in_hours:
        expires_at = datetime.now(timezone.utc) + timedelta(hours=body.expires_in_hours)

    invite = ClubInvite(
        club_id=club_id,
        code=secrets.token_urlsafe(16)[:24],
        created_by=user.id,
        max_uses=body.max_uses,
        expires_at=expires_at,
    )
    db.add(invite)
    await db.commit()
    await db.refresh(invite)
    base_url = str(request.base_url).rstrip("/")
    return _invite_out(invite, base_url)


@router.get("/{club_id}/invites", response_model=list[InviteOut])
async def list_invites(
    club_id: uuid.UUID,
    request: Request,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can view invites")
    result = await db.execute(
        select(ClubInvite).where(ClubInvite.club_id == club_id, ClubInvite.active == True)
    )
    invites = result.scalars().all()
    base_url = str(request.base_url).rstrip("/")
    return [_invite_out(inv, base_url) for inv in invites]


@router.delete("/{club_id}/invites/{invite_id}", status_code=status.HTTP_204_NO_CONTENT)
async def deactivate_invite(
    club_id: uuid.UUID,
    invite_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can deactivate invites")
    result = await db.execute(
        select(ClubInvite).where(ClubInvite.id == invite_id, ClubInvite.club_id == club_id)
    )
    invite = result.scalar_one_or_none()
    if not invite:
        raise NotFoundError("Invite not found")
    invite.active = False
    await db.commit()


# ── Member management ─────────────────────────────────────────────────


@router.post("/{club_id}/members/{user_id}/promote", status_code=status.HTTP_200_OK)
async def promote_member(
    club_id: uuid.UUID,
    user_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    caller = await _get_club_member(db, club_id, user.id)
    if caller.role != "owner":
        raise AuthError("Only the club owner can promote members")
    target = await _get_club_member(db, club_id, user_id)
    if target.role == "owner":
        raise ValidationError("Cannot promote the owner")
    target.role = "admin"
    await db.commit()
    return {"detail": "Member promoted to admin"}


@router.post("/{club_id}/members/{user_id}/demote", status_code=status.HTTP_200_OK)
async def demote_member(
    club_id: uuid.UUID,
    user_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    caller = await _get_club_member(db, club_id, user.id)
    if caller.role != "owner":
        raise AuthError("Only the club owner can demote members")
    target = await _get_club_member(db, club_id, user_id)
    if target.role == "owner":
        raise ValidationError("Cannot demote the owner")
    target.role = "member"
    await db.commit()
    return {"detail": "Member demoted to member"}


@router.delete("/{club_id}/members/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def remove_member(
    club_id: uuid.UUID,
    user_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    caller = await _get_club_member(db, club_id, user.id)
    is_self = user.id == user_id

    if is_self:
        if caller.role == "owner":
            raise ValidationError("Owner cannot leave the club. Transfer ownership or delete the club.")
        await db.delete(caller)
        await db.commit()
        return

    # Removing someone else
    if caller.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can remove members")
    target = await _get_club_member(db, club_id, user_id)
    if target.role == "owner":
        raise ValidationError("Cannot remove the club owner")
    if target.role == "admin" and caller.role != "owner":
        raise AuthError("Only the owner can remove admins")
    await db.delete(target)
    await db.commit()


# ── Leaderboard ───────────────────────────────────────────────────────


@router.get("/{club_id}/leaderboard", response_model=list[LeaderboardOut])
async def get_leaderboard(
    club_id: uuid.UUID,
    template_id: uuid.UUID | None = Query(None),
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    member_ids = await _get_member_user_ids(db, club_id)
    if not member_ids:
        return []

    # Get all completed sessions for club members, optionally filtered by template
    query = (
        select(ScoringSession)
        .where(
            ScoringSession.user_id.in_(member_ids),
            ScoringSession.status == "completed",
        )
    )
    if template_id:
        query = query.where(ScoringSession.template_id == template_id)

    result = await db.execute(query)
    sessions = result.scalars().all()

    # Group by template, find best per user
    from collections import defaultdict
    by_template: dict[uuid.UUID, dict[uuid.UUID, ScoringSession]] = defaultdict(dict)
    for s in sessions:
        tid = s.template_id
        uid = s.user_id
        if uid not in by_template[tid] or s.total_score > by_template[tid][uid].total_score:
            by_template[tid][uid] = s

    # Build user lookup
    user_map: dict[uuid.UUID, User] = {}
    for m_id in member_ids:
        if m_id not in user_map:
            r = await db.execute(select(User).where(User.id == m_id))
            u = r.scalar_one_or_none()
            if u:
                user_map[m_id] = u

    leaderboards = []
    for tid, user_sessions in by_template.items():
        entries = []
        for uid, s in user_sessions.items():
            u = user_map.get(uid)
            if not u:
                continue
            entries.append(LeaderboardEntry(
                user_id=uid,
                username=u.username,
                display_name=u.display_name,
                avatar=u.avatar,
                best_score=s.total_score,
                best_x_count=s.total_x_count,
                session_id=s.id,
                achieved_at=s.completed_at or s.started_at,
            ))
        entries.sort(key=lambda e: e.best_score, reverse=True)
        template_name = "Unknown"
        if s.template:
            template_name = s.template.name
        leaderboards.append(LeaderboardOut(
            template_id=tid,
            template_name=template_name,
            entries=entries,
        ))
    leaderboards.sort(key=lambda lb: lb.template_name)
    return leaderboards


# ── Activity feed ─────────────────────────────────────────────────────


@router.get("/{club_id}/activity", response_model=list[ActivityItem])
async def get_activity(
    club_id: uuid.UUID,
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    member_ids = await _get_member_user_ids(db, club_id)
    if not member_ids:
        return []

    # Build user lookup
    user_map: dict[uuid.UUID, User] = {}
    for m_id in member_ids:
        r = await db.execute(select(User).where(User.id == m_id))
        u = r.scalar_one_or_none()
        if u:
            user_map[m_id] = u

    items: list[ActivityItem] = []

    # Completed sessions
    result = await db.execute(
        select(ScoringSession)
        .where(
            ScoringSession.user_id.in_(member_ids),
            ScoringSession.status == "completed",
        )
        .order_by(ScoringSession.completed_at.desc())
        .limit(limit + offset)
    )
    for s in result.scalars().all():
        u = user_map.get(s.user_id)
        if not u:
            continue
        items.append(ActivityItem(
            type="session_completed",
            user_id=s.user_id,
            username=u.username,
            display_name=u.display_name,
            avatar=u.avatar,
            template_name=s.template.name if s.template else "Unknown",
            score=s.total_score,
            x_count=s.total_x_count,
            session_id=s.id,
            occurred_at=s.completed_at or s.started_at,
        ))

    # Personal records
    result = await db.execute(
        select(PersonalRecord)
        .where(PersonalRecord.user_id.in_(member_ids))
        .order_by(PersonalRecord.achieved_at.desc())
        .limit(limit + offset)
    )
    for pr in result.scalars().all():
        u = user_map.get(pr.user_id)
        if not u:
            continue
        items.append(ActivityItem(
            type="personal_record",
            user_id=pr.user_id,
            username=u.username,
            display_name=u.display_name,
            avatar=u.avatar,
            template_name=pr.template.name if pr.template else "Unknown",
            score=pr.score,
            x_count=0,
            session_id=pr.session_id,
            occurred_at=pr.achieved_at,
        ))

    items.sort(key=lambda i: i.occurred_at, reverse=True)
    return items[offset : offset + limit]


# ── Events ────────────────────────────────────────────────────────────


@router.post("/{club_id}/events", response_model=EventOut, status_code=status.HTTP_201_CREATED)
async def create_event(
    club_id: uuid.UUID,
    body: EventCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can create events")
    event = ClubEvent(
        club_id=club_id,
        name=body.name,
        description=body.description,
        template_id=body.template_id,
        event_date=body.event_date,
        location=body.location,
        created_by=user.id,
    )
    db.add(event)
    await db.commit()
    await db.refresh(event)
    return await _event_out(event, db)


@router.get("/{club_id}/events", response_model=list[EventOut])
async def list_events(
    club_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    result = await db.execute(
        select(ClubEvent)
        .where(ClubEvent.club_id == club_id)
        .order_by(ClubEvent.event_date.desc())
    )
    events = result.scalars().all()
    return [await _event_out(e, db) for e in events]


@router.get("/{club_id}/events/{event_id}", response_model=EventOut)
async def get_event(
    club_id: uuid.UUID,
    event_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    result = await db.execute(
        select(ClubEvent).where(ClubEvent.id == event_id, ClubEvent.club_id == club_id)
    )
    event = result.scalar_one_or_none()
    if not event:
        raise NotFoundError("Event not found")
    return await _event_out(event, db)


@router.patch("/{club_id}/events/{event_id}", response_model=EventOut)
async def update_event(
    club_id: uuid.UUID,
    event_id: uuid.UUID,
    body: EventUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can update events")
    result = await db.execute(
        select(ClubEvent).where(ClubEvent.id == event_id, ClubEvent.club_id == club_id)
    )
    event = result.scalar_one_or_none()
    if not event:
        raise NotFoundError("Event not found")
    if body.name is not None:
        event.name = body.name
    if body.description is not None:
        event.description = body.description
    if body.event_date is not None:
        event.event_date = body.event_date
    if body.location is not None:
        event.location = body.location
    await db.commit()
    await db.refresh(event)
    return await _event_out(event, db)


@router.delete("/{club_id}/events/{event_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_event(
    club_id: uuid.UUID,
    event_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can delete events")
    result = await db.execute(
        select(ClubEvent).where(ClubEvent.id == event_id, ClubEvent.club_id == club_id)
    )
    event = result.scalar_one_or_none()
    if not event:
        raise NotFoundError("Event not found")
    await db.delete(event)
    await db.commit()


@router.post("/{club_id}/events/{event_id}/rsvp", response_model=EventOut)
async def rsvp_event(
    club_id: uuid.UUID,
    event_id: uuid.UUID,
    body: EventRSVP,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    result = await db.execute(
        select(ClubEvent).where(ClubEvent.id == event_id, ClubEvent.club_id == club_id)
    )
    event = result.scalar_one_or_none()
    if not event:
        raise NotFoundError("Event not found")

    # Check existing RSVP
    result = await db.execute(
        select(ClubEventParticipant).where(
            ClubEventParticipant.event_id == event_id,
            ClubEventParticipant.user_id == user.id,
        )
    )
    existing = result.scalar_one_or_none()
    if existing:
        existing.status = body.status
        existing.rsvp_at = datetime.now(timezone.utc)
    else:
        participant = ClubEventParticipant(
            event_id=event_id,
            user_id=user.id,
            status=body.status,
        )
        db.add(participant)
    await db.commit()
    await db.refresh(event)
    return await _event_out(event, db)


def _team_member_out(tm) -> TeamMemberOut:
    return TeamMemberOut(
        user_id=tm.user_id,
        username=tm.user.username,
        display_name=tm.user.display_name,
        avatar=tm.user.avatar,
        joined_at=tm.joined_at,
    )


def _team_out(team: ClubTeam) -> TeamOut:
    return TeamOut(
        id=team.id,
        club_id=team.club_id,
        name=team.name,
        description=team.description,
        leader=TeamMemberOut(
            user_id=team.leader.id,
            username=team.leader.username,
            display_name=team.leader.display_name,
            avatar=team.leader.avatar,
            joined_at=team.created_at,
        ),
        member_count=len(team.members),
        created_at=team.created_at,
    )


def _team_detail_out(team: ClubTeam) -> TeamDetailOut:
    return TeamDetailOut(
        id=team.id,
        club_id=team.club_id,
        name=team.name,
        description=team.description,
        leader=TeamMemberOut(
            user_id=team.leader.id,
            username=team.leader.username,
            display_name=team.leader.display_name,
            avatar=team.leader.avatar,
            joined_at=team.created_at,
        ),
        member_count=len(team.members),
        created_at=team.created_at,
        members=[_team_member_out(m) for m in team.members],
    )


async def _event_out(event: ClubEvent, db: AsyncSession) -> EventOut:
    now = datetime.now(timezone.utc)
    is_past = event.event_date < now

    participants = []
    for p in event.participants:
        r = await db.execute(select(User).where(User.id == p.user_id))
        u = r.scalar_one_or_none()
        if not u:
            continue

        score = None
        x_count = None
        session_id = None

        # If event is past and participant was "going", find their best session on that date
        if is_past and p.status == "going":
            event_date = event.event_date.date()
            r2 = await db.execute(
                select(ScoringSession)
                .where(
                    ScoringSession.user_id == p.user_id,
                    ScoringSession.template_id == event.template_id,
                    ScoringSession.status == "completed",
                    func.date(ScoringSession.completed_at) == event_date,
                )
                .order_by(ScoringSession.total_score.desc())
                .limit(1)
            )
            session = r2.scalar_one_or_none()
            if session:
                score = session.total_score
                x_count = session.total_x_count
                session_id = session.id

        participants.append(EventParticipantOut(
            user_id=p.user_id,
            username=u.username,
            display_name=u.display_name,
            avatar=u.avatar,
            status=p.status,
            score=score,
            x_count=x_count,
            session_id=session_id,
        ))

    return EventOut(
        id=event.id,
        club_id=event.club_id,
        name=event.name,
        description=event.description,
        template_id=event.template_id,
        template_name=event.template.name if event.template else None,
        event_date=event.event_date,
        location=event.location,
        created_by=event.created_by,
        participants=participants,
        created_at=event.created_at,
    )


# ── Teams ────────────────────────────────────────────────────────────


@router.post("/{club_id}/teams", response_model=TeamOut, status_code=status.HTTP_201_CREATED)
async def create_team(
    club_id: uuid.UUID,
    body: TeamCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can create teams")

    # Validate leader is a club member
    leader_member = await db.execute(
        select(ClubMember).where(ClubMember.club_id == club_id, ClubMember.user_id == body.leader_id)
    )
    if not leader_member.scalar_one_or_none():
        raise ValidationError("Leader must be an existing club member")

    team = ClubTeam(
        club_id=club_id,
        name=body.name,
        description=body.description,
        leader_id=body.leader_id,
    )
    db.add(team)
    await db.commit()
    await db.refresh(team)
    return _team_out(team)


@router.get("/{club_id}/teams", response_model=list[TeamOut])
async def list_teams(
    club_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    result = await db.execute(
        select(ClubTeam).where(ClubTeam.club_id == club_id).order_by(ClubTeam.name)
    )
    teams = result.scalars().all()
    return [_team_out(t) for t in teams]


@router.get("/{club_id}/teams/{team_id}", response_model=TeamDetailOut)
async def get_team(
    club_id: uuid.UUID,
    team_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_club_member(db, club_id, user.id)
    result = await db.execute(
        select(ClubTeam).where(ClubTeam.id == team_id, ClubTeam.club_id == club_id)
    )
    team = result.scalar_one_or_none()
    if not team:
        raise NotFoundError("Team not found")
    return _team_detail_out(team)


@router.patch("/{club_id}/teams/{team_id}", response_model=TeamOut)
async def update_team(
    club_id: uuid.UUID,
    team_id: uuid.UUID,
    body: TeamUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can update teams")

    result = await db.execute(
        select(ClubTeam).where(ClubTeam.id == team_id, ClubTeam.club_id == club_id)
    )
    team = result.scalar_one_or_none()
    if not team:
        raise NotFoundError("Team not found")

    if body.name is not None:
        team.name = body.name
    if body.description is not None:
        team.description = body.description
    if body.leader_id is not None:
        leader_member = await db.execute(
            select(ClubMember).where(ClubMember.club_id == club_id, ClubMember.user_id == body.leader_id)
        )
        if not leader_member.scalar_one_or_none():
            raise ValidationError("Leader must be an existing club member")
        team.leader_id = body.leader_id

    await db.commit()
    await db.refresh(team)
    return _team_out(team)


@router.delete("/{club_id}/teams/{team_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_team(
    club_id: uuid.UUID,
    team_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    member = await _get_club_member(db, club_id, user.id)
    if member.role not in ("owner", "admin"):
        raise AuthError("Only owners and admins can delete teams")

    result = await db.execute(
        select(ClubTeam).where(ClubTeam.id == team_id, ClubTeam.club_id == club_id)
    )
    team = result.scalar_one_or_none()
    if not team:
        raise NotFoundError("Team not found")
    await db.delete(team)
    await db.commit()


@router.post("/{club_id}/teams/{team_id}/members/{user_id}", status_code=status.HTTP_201_CREATED)
async def add_team_member(
    club_id: uuid.UUID,
    team_id: uuid.UUID,
    user_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    caller = await _get_club_member(db, club_id, user.id)

    result = await db.execute(
        select(ClubTeam).where(ClubTeam.id == team_id, ClubTeam.club_id == club_id)
    )
    team = result.scalar_one_or_none()
    if not team:
        raise NotFoundError("Team not found")

    # Permission check: owner/admin or team leader
    if caller.role not in ("owner", "admin") and team.leader_id != user.id:
        raise AuthError("Only owners, admins, or the team leader can manage team members")

    # Validate target is a club member
    target_member = await db.execute(
        select(ClubMember).where(ClubMember.club_id == club_id, ClubMember.user_id == user_id)
    )
    if not target_member.scalar_one_or_none():
        raise ValidationError("User must be an existing club member")

    # Check not already on team
    existing = await db.execute(
        select(ClubTeamMember).where(ClubTeamMember.team_id == team_id, ClubTeamMember.user_id == user_id)
    )
    if existing.scalar_one_or_none():
        raise ConflictError("User is already a member of this team")

    tm = ClubTeamMember(team_id=team_id, user_id=user_id)
    db.add(tm)
    await db.commit()
    return {"detail": "Member added to team"}


@router.delete("/{club_id}/teams/{team_id}/members/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def remove_team_member(
    club_id: uuid.UUID,
    team_id: uuid.UUID,
    user_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    caller = await _get_club_member(db, club_id, user.id)

    result = await db.execute(
        select(ClubTeam).where(ClubTeam.id == team_id, ClubTeam.club_id == club_id)
    )
    team = result.scalar_one_or_none()
    if not team:
        raise NotFoundError("Team not found")

    # Permission check: owner/admin or team leader
    if caller.role not in ("owner", "admin") and team.leader_id != user.id:
        raise AuthError("Only owners, admins, or the team leader can manage team members")

    existing = await db.execute(
        select(ClubTeamMember).where(ClubTeamMember.team_id == team_id, ClubTeamMember.user_id == user_id)
    )
    tm = existing.scalar_one_or_none()
    if not tm:
        raise NotFoundError("User is not a member of this team")
    await db.delete(tm)
    await db.commit()
