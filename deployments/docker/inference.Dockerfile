FROM ghcr.io/astral-sh/uv:python3.11-bookworm-slim

WORKDIR /app

COPY services/inference/pyproject.toml services/inference/uv.lock ./
RUN uv sync --frozen --extra model

COPY services/gen/python /app/gen/python
COPY services/inference/server.py services/inference/model.py ./

ENV PYTHONPATH=/app/gen/python
ENV INFERENCE_BIND=[::]:50057

EXPOSE 50057

CMD ["uv", "run", "python", "server.py"]
