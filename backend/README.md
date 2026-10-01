# CargoPrener API

Go REST API for CargoPrener. Auth is implemented. The rest of the schema is migrated so later endpoints can use it.

The frontend does not have an API client yet. There is no `src/api` folder and no `VITE_USE_MOCKS` flag. Landing types also differ from these tables: the UI puts `moq` on the product and uses `location` / `moderated` / `price`, while the database stores `orders.moq`, `suppliers.region`, `suppliers.is_verified`, and `price_tiers.unit_price`. Collected quantity and participant count are not columns; they come from `order_participations`.

## Run with Docker

From the repository root:

```bash
cp .env.example .env
```

Edit `.env` and replace the placeholder secrets. `DATABASE_URL` must use the host `postgres` inside Compose, and the password must match `POSTGRES_PASSWORD`.

```bash
docker compose up --build
```

- API: http://localhost:8080
- Landing page: http://localhost:8081
- Postgres: localhost:5432

The API applies SQL migrations on startup, then creates the admin user when `ADMIN_EMAIL` and `ADMIN_PASSWORD` are both set. The admin password is never stored in the repository. If that email already exists, startup promotes it to `admin` and leaves the password unchanged.

## Run the API locally

Install Go 1.27 or newer. Start Postgres (Compose can run only the database):

```bash
docker compose up postgres
```

From the `backend` folder, with Postgres already running:

```powershell
$env:DATABASE_URL = "postgres://cargoprener:replace-with-a-database-password@localhost:5432/cargoprener?sslmode=disable"
$env:JWT_SECRET = "replace-with-a-long-random-string"
$env:CORS_ORIGIN = "http://localhost:5173"
$env:COOKIE_SECURE = "false"
$env:ADMIN_EMAIL = "admin@cargoprener.local"
$env:ADMIN_PASSWORD = "replace-with-an-admin-password"
go run ./cmd/api
```

The process also loads `.env` from `backend` and from the repository root. Migrations run on startup.

To reapply migrations by hand, restart the API. golang-migrate records applied versions in `schema_migrations`.

## Tests

```bash
go test ./...
```

Auth handler tests use an in-memory user store. They do not need Postgres.

## Environment

| Variable | Purpose |
| --- | --- |
| `PORT` | API port. Default `8080`. |
| `DATABASE_URL` | Postgres URL. Required. |
| `JWT_SECRET` | HMAC secret, at least 16 characters. Required. |
| `CORS_ORIGIN` | Frontend origin allowed to send cookies. Default `http://localhost:5173`. |
| `COOKIE_SECURE` | Set `true` when the site is served over HTTPS. |
| `ADMIN_EMAIL` | Admin account created on startup if it does not exist. |
| `ADMIN_PASSWORD` | Password for that new admin account. |
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Used by the Compose Postgres service. |

## Endpoints

Session cookie name: `session`. It is `HttpOnly`, `SameSite=Lax`, and `Secure` when `COOKIE_SECURE=true`. Lifetime is 24 hours. The token is not returned in JSON.

### `GET /health`

`200`

```json
{"status":"ok"}
```

`503` when Postgres cannot be reached:

```json
{"message":"Database unavailable"}
```

### `POST /auth/register`

```json
{
  "businessName": "Harbor Shop",
  "contactName": "Ada Lopez",
  "email": "ada@harbor.example",
  "password": "long-enough",
  "region": "Grand Rapids, MI"
}
```

`201` sets the session cookie and returns:

```json
{
  "id": "uuid",
  "businessName": "Harbor Shop",
  "contactName": "Ada Lopez",
  "email": "ada@harbor.example",
  "region": "Grand Rapids, MI",
  "role": "business",
  "createdAt": "2026-09-30T12:00:00Z"
}
```

`422` when a field is missing, the email is invalid, or the password is shorter than 8 characters:

```json
{
  "message": "Check the highlighted fields",
  "fieldErrors": {"password": "Password must be at least 8 characters"}
}
```

`409` when the email is already registered:

```json
{
  "message": "Email is already registered",
  "fieldErrors": {"email": "An account with this email already exists"}
}
```

### `POST /auth/login`

```json
{"email": "ada@harbor.example", "password": "long-enough"}
```

`200` with the same user object and a new session cookie. Wrong email and wrong password both return `401`:

```json
{"message": "Invalid email or password"}
```

### `POST /auth/logout`

`204` and the session cookie is cleared.

### `GET /auth/me`

`200` with the user object when the session cookie is valid. Otherwise `401`:

```json
{"message": "Authentication required"}
```

`/auth/login` and `/auth/register` are limited to 10 requests per minute per address.

## Switching the frontend to this API

The landing page still has no auth screens and no fetch client, so an env flag alone cannot switch it yet. When that client is added:

1. Point it at `http://localhost:8080`.
2. Send cookies on every auth request: `fetch(url, { credentials: "include" })`.
3. Keep `CORS_ORIGIN` equal to the frontend origin (`http://localhost:5173` for Vite, `http://localhost:8081` for the Compose nginx build).
4. Leave `COOKIE_SECURE=false` on local HTTP.

`localhost` ports are the same site, so the `SameSite=Lax` cookie is sent on those credentialed requests.
