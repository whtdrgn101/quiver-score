FROM python:3.14-slim

WORKDIR /code

COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev

COPY alembic.ini ./
COPY alembic/ alembic/
COPY app/ app/

# Migration-only image: the deploy runs this as a Cloud Run *job*. No server.
CMD ["uv", "run", "alembic", "upgrade", "head"]
