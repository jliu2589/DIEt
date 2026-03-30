# Nutrition Backend (Go)

## What this service does
This backend powers meal logging and nutrition tracking with:
- meal ingestion + analysis (`POST /v1/meals`),
- chat intent routing (`POST /v1/chat`),
- meal history/detail/edit/delete,
- daily summaries + recommendations,
- user settings, weight tracking, trends,
- optional Telegram webhook integration.

## Tech stack
- Go + Gin
- PostgreSQL + pgx
- OpenAI (optional fallback)
- Telegram Bot API (optional integration)

## Configuration
### Always required
- `PORT`
- `DATABASE_URL`

### Optional feature flags
- `ENABLE_OPENAI_FALLBACK` (default: enabled only when `OPENAI_API_KEY` is set)
- `ENABLE_TELEGRAM_INTEGRATION` (default: enabled only when all Telegram vars are set)

### Required when OpenAI fallback is enabled
- `OPENAI_API_KEY`

### Required when Telegram integration is enabled
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_WEBHOOK_SECRET_PATH`
- `TELEGRAM_WEBHOOK_SECRET_TOKEN`

> See `.env.example` for a starter template.

## Local setup
```bash
cp .env.example .env
# edit .env
make migrate-up
make run
```

Server starts on `:$PORT`.

## Auth in local/dev
The API uses demo-subject middleware. You can pass a user via:
- `X-Demo-User-ID: <user-id>` header, or
- `user_id` in query/body on endpoints that support it.

## Current HTTP routes
- `GET /health`
- `POST /v1/meals`
- `GET /v1/meals/:mealEventID?user_id=<id>`
- `PATCH /v1/meals/:mealEventID`
- `DELETE /v1/meals/:mealEventID?user_id=<id>`
- `GET /v1/meals/recent?user_id=<id>&limit=20`
- `PATCH /v1/meals/:mealEventID/time`
- `POST /v1/chat`
- `GET /v1/daily-summary?user_id=<id>&date=YYYY-MM-DD`
- `GET /v1/recommendations?user_id=<id>`
- `GET /v1/dashboard/today?user_id=<id>`
- `GET /v1/foods/search?q=<text>&limit=20`
- `GET /v1/reusable-meals?limit=20`
- `GET /v1/settings?user_id=<id>`
- `PUT /v1/settings`
- `POST /v1/weight`
- `GET /v1/weight/latest?user_id=<id>`
- `GET /v1/weight/recent?user_id=<id>&limit=30`
- `GET /v1/trends?user_id=<id>&range=7d|30d|90d|1y`
- `GET /v1/me?user_id=<id>`
- `POST /v1/integrations/telegram/webhook/:secretPath` (only when Telegram integration enabled)

---

## Contract examples

### `POST /v1/meals`
Request:
```bash
curl -X POST "http://localhost:8080/v1/meals" \
  -H "Content-Type: application/json" \
  -H "X-Demo-User-ID: demo-user" \
  -d '{
    "source": "web",
    "raw_text": "chicken rice bowl",
    "eaten_at": "2026-03-30T12:30:00Z"
  }'
```

Response shape:
```json
{
  "intent": "meal_log",
  "logged": true,
  "message": "Logged your meal.",
  "item": {
    "meal_event_id": 123,
    "canonical_name": "chicken rice bowl",
    "logged_at": "2026-03-30T12:31:05Z",
    "eaten_at": "2026-03-30T12:30:00Z",
    "time_source": "explicit",
    "source": "web",
    "confidence_score": 0.91,
    "calories_kcal": 620,
    "protein_g": 38,
    "carbohydrate_g": 62,
    "fat_g": 24
  }
}
```

### `GET /v1/meals/:mealEventID`
Returns one meal detail (analysis + time metadata):
```json
{
  "meal_event_id": 123,
  "canonical_name": "chicken rice bowl",
  "logged_at": "2026-03-30T12:31:05Z",
  "eaten_at": "2026-03-30T12:30:00Z",
  "time_source": "explicit",
  "source": "web",
  "raw_text": "chicken rice bowl",
  "assumptions": ["standard serving"],
  "items": [{ "name": "chicken", "quantity": 150, "unit": "g" }],
  "calories_kcal": 620,
  "protein_g": 38,
  "carbohydrate_g": 62,
  "fat_g": 24
}
```

### `GET /v1/recommendations`
Recommendation items include source metadata:
- `source`: `meal_memory` | `reusable_meal` | `canonical_food`
- `source_id`: source record id

## OpenAPI
See `docs/openapi.yaml` for the current API contract.
