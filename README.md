# go-chirpy!

A small HTTP API server (and tiny static app) that runs on `http://localhost:8080`.

## Run

This project loads environment variables from `.env` on startup.

```bash
cd go-chirpy

go run .
# or:
go build -o chirpy .
./chirpy
```

**Required env vars** (see `.env` for local defaults):
- `DB_URL` (Postgres connection string)
- `JWT_SECRET` (HMAC secret for access tokens)
- `PLATFORM` (set to `dev` to enable `POST /admin/reset`)
- `POLKA_KEY` (API key for `POST /api/polka/webhooks`)

## Base URL

- `http://localhost:8080`

## Typical flow

1) Create an account: `POST /api/users`
2) Log in to get tokens: `POST /api/login`
3) Use the **access token** as `Authorization: Bearer <token>` for protected routes (create chirp, update user, delete chirp)
4) Use the **refresh token** with `POST /api/refresh` to get a new access token
5) Optionally revoke a refresh token with `POST /api/revoke`

---

## Auth

### `POST /api/login`

Request:
```bash
curl -sS -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@mail.com","password":"test"}'
```

Response (`200`):
```json
{
  "id": "...",
  "created_at": "...",
  "updated_at": "...",
  "email": "...",
  "token": "<access token>",
  "refresh_token": "<refresh token>",
  "is_chirpy_red": false
}
```

Errors:
- `401` for incorrect email/password

### `POST /api/refresh`

Uses the **refresh token** as a Bearer token.

```bash
curl -sS -X POST http://localhost:8080/api/refresh \
  -H 'Authorization: Bearer <refresh_token>'
```

Response (`200`):
```json
{ "token": "<new access token>" }
```

### `POST /api/revoke`

Revokes a refresh token.

```bash
curl -i -X POST http://localhost:8080/api/revoke \
  -H 'Authorization: Bearer <refresh_token>'
```

Response: `204 No Content`

---

## Users

### `POST /api/users`

Create a new user.

```bash
curl -sS -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@mail.com","password":"test"}'
```

Response (`201`):
```json
{
  "id": "...",
  "created_at": "...",
  "updated_at": "...",
  "email": "...",
  "is_chirpy_red": false
}
```

### `PUT /api/users`

Update the logged-in user’s email/password.

```bash
curl -sS -X PUT http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <access_token>' \
  -d '{"email":"new@mail.com","password":"new-password"}'
```

Response (`200`):
```json
{
  "id": "...",
  "created_at": "...",
  "updated_at": "...",
  "email": "...",
  "is_chirpy_red": false
}
```

---

## Chirps

A chirp body must be **<= 140 characters**. Some words are automatically censored.

### `GET /api/chirps`

List all chirps:
```bash
curl -sS http://localhost:8080/api/chirps
```

Filter by author (UUID):
```bash
curl -sS 'http://localhost:8080/api/chirps?author_id=<user_uuid>'
```

Response (`200`): JSON array of chirps.

### `GET /api/chirps/{chirpID}`

```bash
curl -sS http://localhost:8080/api/chirps/<chirp_uuid>
```

Response (`200`):
```json
{
  "id": "...",
  "created_at": "...",
  "updated_at": "...",
  "body": "...",
  "user_id": "..."
}
```

### `POST /api/chirps`

Create a chirp (requires access token):

```bash
curl -sS -X POST http://localhost:8080/api/chirps \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <access_token>' \
  -d '{"body":"hello chirpy"}'
```

Response (`201`): the created chirp.

### `DELETE /api/chirps/{chirpID}`

Delete your own chirp (requires access token):

```bash
curl -i -X DELETE http://localhost:8080/api/chirps/<chirp_uuid> \
  -H 'Authorization: Bearer <access_token>'
```

Response: `204 No Content`

---

## Admin

### `GET /admin/metrics`

Returns a small HTML page showing how many times the static app was visited.

```bash
curl -sS http://localhost:8080/admin/metrics
```

### `POST /admin/reset`

**Dev-only**: requires `PLATFORM=dev`.
- Deletes all users
- Resets the static app hit counter

```bash
curl -i -X POST http://localhost:8080/admin/reset
```

---

## Webhooks

### `POST /api/polka/webhooks`

Validates an API key using:

- `Authorization: ApiKey <POLKA_KEY>`

Example request:
```bash
curl -i -X POST http://localhost:8080/api/polka/webhooks \
  -H 'Content-Type: application/json' \
  -H 'Authorization: ApiKey <POLKA_KEY>' \
  -d '{"event":"user.upgraded","data":{"user_id":"<user_uuid>"}}'
```

Response:
- `204 No Content` for supported/unsupported events

---

## Other

### `GET /api/healthz`

```bash
curl -sS http://localhost:8080/api/healthz
```

Response (`200`): `OK`

### Static app

- `GET /app/` serves files from the `go-chirpy` directory (e.g. `index.html`).
- Visiting `/app/*` increments the “hits” counter shown in `/admin/metrics`.
