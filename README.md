# Nutrition Backend (Go)

## Project Purpose
This service is a nutrition tracking backend that:
- accepts meal text input from API clients and Telegram,
- analyzes meals (with OpenAI),
- caches prior analyses using meal fingerprinting,
- stores meal events and nutrition analysis in Postgres,
- maintains per-user daily nutrition totals.

## Stack
- **Go**
- **Gin** (HTTP router)
- **pgx** (Postgres driver/pool)
- **PostgreSQL**
- **Telegram Bot API**
- **OpenAI API**

## Required Environment Variables
Set these before starting the API:

- `PORT`
- `DATABASE_URL`
- `OPENAI_API_KEY`
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_WEBHOOK_SECRET_PATH`
- `TELEGRAM_WEBHOOK_SECRET_TOKEN`

See `.env.example` for a starter template.

## Run Locally
### 1) Prepare environment
```bash
cp .env.example .env
# edit .env with real values
```

### 2) Start PostgreSQL and apply migrations
Make sure your Postgres instance is reachable via `DATABASE_URL`, then run:

```bash
make migrate-up
```

> Requires the `migrate` CLI (`golang-migrate`) installed locally.

### 3) Start API
```bash
make run
```

Server starts on `:$PORT`.

## HTTP Routes
- `GET /health`
- `POST /v1/meals`
- `GET /v1/meals/recent?user_id=<id>&limit=20`
- `GET /v1/daily-summary?user_id=<id>&date=YYYY-MM-DD`
- `GET /v1/settings?user_id=<id>`
- `PUT /v1/settings`
- `POST /v1/weight`
- `GET /v1/weight/latest?user_id=<id>`
- `GET /v1/weight/recent?user_id=<id>&limit=30`
- `POST /v1/integrations/telegram/webhook/:secretPath`

## Example: `POST /v1/meals`

Request:
```bash
curl -X POST "http://localhost:8080/v1/meals" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "source": "web",
    "raw_text": "chicken rice bowl and avocado",
    "eaten_at": "2026-03-27T12:30:00Z"
  }'
```

Example response:
```json
{
  "meal_event_id": 101,
  "processed_from": "openai",
  "canonical_name": "chicken rice bowl",
  "nutrition": {
    "calories_kcal": 620,
    "protein_g": 38,
    "carbohydrate_g": 62,
    "fat_g": 24,
    "fiber_g": 9,
    "sugars_g": 4,
    "saturated_fat_g": 4,
    "sodium_mg": 720,
    "potassium_mg": 950,
    "calcium_mg": 80,
    "magnesium_mg": 110,
    "iron_mg": 3.5,
    "zinc_mg": 2.1,
    "vitamin_d_mcg": 0,
    "vitamin_b12_mcg": 0.4,
    "folate_b9_mcg": 120,
    "vitamin_c_mg": 14
  },
  "confidence_score": 0.86
}
```

## Telegram Webhook Setup
1. Configure your env vars:
   - `TELEGRAM_BOT_TOKEN`
   - `TELEGRAM_WEBHOOK_SECRET_PATH`
   - `TELEGRAM_WEBHOOK_SECRET_TOKEN`

2. Build webhook URL:
   ```text
   https://<your-domain>/v1/integrations/telegram/webhook/<TELEGRAM_WEBHOOK_SECRET_PATH>
   ```

3. Register webhook with Telegram:
   ```bash
   curl -X POST "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/setWebhook" \
     -H "Content-Type: application/json" \
     -d '{
       "url": "https://<your-domain>/v1/integrations/telegram/webhook/<TELEGRAM_WEBHOOK_SECRET_PATH>",
       "secret_token": "<TELEGRAM_WEBHOOK_SECRET_TOKEN>"
     }'
   ```

4. Telegram will send updates to your webhook endpoint. The handler validates:
   - URL path secret,
   - `X-Telegram-Bot-Api-Secret-Token` header.

## Example Webhook Flow (V1)
1. User sends text message to Telegram bot:
   - e.g. `"2 eggs and toast"`
2. Telegram sends webhook update to:
   - `POST /v1/integrations/telegram/webhook/:secretPath`
3. Backend validates webhook secret path and header token.
4. Backend maps Telegram user to app user:
   - `user_id = "telegram:<telegram_user_id>"`
5. Backend processes meal text:
   - fingerprint + cache lookup,
   - OpenAI analysis on cache miss,
   - persist event/analysis/memory,
   - update daily summary.
6. Backend replies to Telegram chat with meal name and key macros:
   - calories, protein, carbs, fat.

---
If you want, next step is adding Docker Compose for local Postgres + API startup in one command.
