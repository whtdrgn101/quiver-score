from fastapi import APIRouter

from app.api.v1.auth import router as auth_router
from app.api.v1.users import router as users_router
from app.api.v1.rounds import router as rounds_router
from app.api.v1.scoring import router as scoring_router
from app.api.v1.equipment import router as equipment_router
from app.api.v1.setups import router as setups_router
from app.api.v1.sharing import router as sharing_router
from app.api.v1.clubs import router as clubs_router
from app.api.v1.notifications import router as notifications_router
from app.api.v1.classifications import router as classifications_router
from app.api.v1.sight_marks import router as sight_marks_router
from app.api.v1.coaching import router as coaching_router
from app.api.v1.social import router as social_router

api_router = APIRouter(prefix="/api/v1")
api_router.include_router(auth_router)
api_router.include_router(users_router)
api_router.include_router(rounds_router)
api_router.include_router(scoring_router)
api_router.include_router(equipment_router)
api_router.include_router(setups_router)
api_router.include_router(sharing_router)
api_router.include_router(clubs_router)
api_router.include_router(notifications_router)
api_router.include_router(classifications_router)
api_router.include_router(sight_marks_router)
api_router.include_router(coaching_router)
api_router.include_router(social_router)
