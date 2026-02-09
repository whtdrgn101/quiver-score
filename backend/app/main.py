from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import settings
from app.database import async_session
from app.api.v1 import api_router
from app.seed.round_templates import seed_round_templates


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with async_session() as db:
        await seed_round_templates(db)
    yield


app = FastAPI(title="QuiverScore", version="0.1.0", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(api_router)


@app.get("/health")
async def health():
    return {"status": "ok"}
