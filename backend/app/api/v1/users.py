import base64

import httpx
from fastapi import APIRouter, Depends, HTTPException, UploadFile
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.schemas.user import AvatarUrlUpload, UserOut, UserUpdate

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
            resp = await client.get(body.url)
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
