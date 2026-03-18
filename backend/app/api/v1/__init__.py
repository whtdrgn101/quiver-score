from fastapi import APIRouter

from app.api.v1.scoring import router as scoring_router

api_router = APIRouter(prefix="/api/v1")
api_router.include_router(scoring_router)
