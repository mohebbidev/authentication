## Advanced Authentication (Production-Ready) — Full Flow Using Your Current Tables

You currently have:

- `sessions (session_id, user_id, expires_at, revoked_at)`
- `tokens (token_hash PK, expires_at, revoked_at)`
- `refresh_tokens` **INHERITS** from `tokens`
- `verification_tokens` **INHERITS** from `tokens`
- `password_reset_tokens` **INHERITS** from `tokens`

### Key note before we dive:
Postgres `INHERITS` means those tables share columns from `tokens`. But **you’re missing required columns** for rotation/auditing (like token owner columns, issued_at, etc.) unless `tokens` is extended further. For “advanced auth” you can still use your design, but you should adjust it slightly (details below).

I’ll give you the **complete flow** (register/login/refresh/logout/reset/verify), and also the **exact DB operations** and **security controls** to wire.

---

# 0) What you actually need (clear requirements)

## Needed token types
1) **Access token**: short-lived JWT *or* opaque (your choice).  
2) **Refresh tokens**: long-lived, stored server-side as **hashes** (you already have `token_hash` + `revoked_at`).
3) **Email verification token**: one-time, short-lived.
4) **Password reset token**: one-time, short-lived.
5) **Sessions** (optional but useful): you can keep it as “device session” records and bind refresh tokens to a session.

## Needed user controls
- Password hashing: **Argon2id** (or bcrypt as fallback)
- Normalize email: lowercase + trim
- Rate limiting:
  - login endpoint
  - refresh endpoint
  - reset/verify endpoints
- Generic error responses on auth failures
- Rotation for refresh tokens (to prevent replay)

---

# 1) Fix/upgrade your schema for “advanced” correctness

## 1.1 Your `tokens` table is missing ownership + type separation context
Right now `tokens` only stores:
- `token_hash`
- `expires_at`
- `revoked_at`

But for refresh/verify/reset you almost always need:
- who it belongs to (`user_id` or other subject)
- what it is for (type) OR separate tables (you already separated)
- whether it was used (especially one-time tokens)
- rotation lineage (refresh rotation)

### Recommended minimal additions
**A) Add `used_at` to one-time tokens** (verification/reset)  
**B) Add `user_id` and maybe `session_id` for refresh tokens**  
**C) Add `issued_at` for auditing + expiry calculations**
  
Since you already inherit, do this on the parent `tokens` only if all children need it.

#### Option: simplest practical upgrade
Add these columns to `tokens`:
- `user_id UUID`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `used_at TIMESTAMPTZ` (for one-time)
- `reason TEXT` (optional: revoke reason)
- `rotated_from_hash TEXT` (optional for refresh rotation lineage)

But if you want to keep `user_id` only for refresh/verify/reset, then add it on each child instead.

---

## 1.2 Your `sessions` table indexing is too limiting
Right now:
```sql
CREATE UNIQUE INDEX idx_sessions_userid ON sessions (user_id)
```
That makes **only one session per user**. For production you want multiple sessions (devices).

Change to:
- Either remove uniqueness
- Add indexes:
  - `sessions(user_id)`
  - `sessions(session_id)` is already PK
  - optionally `sessions(user_id, expires_at)`

Example:
```sql
DROP INDEX IF EXISTS idx_sessions_userid;

CREATE INDEX IF NOT EXISTS idx_sessions_userid ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
```

---

## 1.3 Refresh token rotation needs a way to revoke the used refresh token
Your `refresh_tokens` has `token_hash PK` so you can:
- look up by hash
- if valid: revoke it and insert the new hash

To detect replay, you should revoke-at-rotation time.

---

# 2) Token generation rules (what is needed vs not)

## Needed
- Generate refresh/verify/reset tokens as **opaque random strings**
  - 256-bit+ randomness (e.g. 32 bytes) encoded base64url
- Store only `hash(token)` in DB
  - Use SHA-256 (or stronger KDF) for hashing tokens
- Never store plaintext tokens in DB

## Not needed
- JWT for refresh tokens (don’t do that for refresh)
- Storing access token in DB (if JWT with short exp)
- `SELECT *` for login hot paths

---

# 3) Endpoint flows (full)

I’ll assume:
- Access token is JWT (short exp) OR opaque—either way, refresh rotates.
- Refresh token is in HttpOnly cookie **or** JSON; your architecture can support either.

## 3.1 Register: `POST /auth/register`
**Input**: email, password, (optional) name  
**Steps**:
1) Validate & normalize email (lowercase trim)
2) Check uniqueness
3) Hash password (Argon2id)
4) Create user (inactive until verified if you require)
5) Create **verification token**:
   - generate random `plain`
   - compute `hash`
   - insert into `verification_tokens` with:
     - `token_hash = hash`
     - `expires_at = now + 15m/1h`
     - `user_id = user.id` (you need this column)
     - `used_at = NULL`
6) Send email with token link

**Response**:
- `201 Created` (or `200 OK`)
- Generic message: “If an account exists, we sent instructions.”

### What to enforce
- Rate limit register requests (especially by IP)
- Do not reveal if email exists

---

## 3.2 Verify email: `POST /auth/verify`
**Input**: token (from link)  
**Steps**:
1) Hash token
2) `SELECT user_id, expires_at, used_at, revoked_at FROM verification_tokens WHERE token_hash = $1`
3) If missing/expired/revoked/used → return generic success or 400 generic
4) Mark as used:
   - `UPDATE verification_tokens SET used_at = now(), revoked_at = now() WHERE token_hash = $1`
5) Activate user:
   - `UPDATE users SET is_active = true, email_verified_at = now() WHERE id = $user_id`

**Response**: `204 No Content` or `200 OK`

### Needed security
- One-time token: `used_at` required
- Rate limit verification endpoint

---

## 3.3 Login: `POST /auth/login`
**Input**: email, password

**Steps**:
1) Rate limit by:
   - IP + email (separately if possible)
2) Normalize email
3) Fetch user:
   - `SELECT id, password_hash, is_active FROM users WHERE email=$1`
4) If user missing OR password mismatch OR not active:
   - return `401` generic
5) Generate:
   - **access token** (short exp, JWT)
   - **refresh token** (opaque random string)
   - (optional) create `sessions` row for the device
     - session_id random UUID
     - expires_at = refresh expiry
6) Store refresh token hash:
   - `INSERT INTO refresh_tokens(token_hash, expires_at, revoked_at, user_id, session_id) VALUES (...)`
   - (fields depend on your final schema)
7) Set cookie:
   - `refresh_token=plain` (HttpOnly, Secure, SameSite=Lax)
8) Return access token (JSON) or set access cookie

**Response**:
- `200 OK`

### Needed security
- Generic failures (no user enumeration)
- Brute force control (rate limiting)
- Consider “account lock” strategy only if you can do it safely

---

## 3.4 Refresh access: `POST /auth/refresh`
This is the core for “advanced”.

**Input**:
- read refresh token from cookie (preferred), else from JSON

**Steps (rotation):**
1) If refresh cookie missing → `401`
2) Hash refresh token
3) `SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash=$1`
4) If not found/expired/revoked → return `401` generic  
5) **Rotate**:
   - `UPDATE refresh_tokens SET revoked_at=now() WHERE token_hash=$1`
   - generate new refresh token `plain2`
   - insert hash2 with new expiry (and link to session if you store `session_id`)
6) Issue new access token
7) Set new refresh cookie (overwrite old cookie)

**Replay handling (advanced)**
If an attacker steals a refresh token and tries to reuse it:
- the second attempt will find it already revoked → treat as suspicious
- you may optionally revoke all refresh tokens for that user (requires an index + query)

---

## 3.5 Logout: `POST /auth/logout`
**Steps**:
1) Hash refresh token from cookie
2) `UPDATE refresh_tokens SET revoked_at=now() WHERE token_hash=$1`
3) Clear refresh cookie
4) (Optional) revoke `sessions` row or mark it ended

**Response**:
- `204 No Content`

---

## 3.6 Password reset request: `POST /auth/password-reset/request`
**Input**: email

**Steps**:
1) Rate limit strongly
2) Normalize email
3) Always respond generically
4) If user exists:
   - create password reset token (one-time):
     - generate opaque random token
     - hash it
     - insert into `password_reset_tokens` with:
       - `user_id`, `expires_at`
       - `used_at = NULL`
   - send email link

**Response**:
- `200 OK` generic

---

## 3.7 Password reset: `POST /auth/password-reset/confirm`
**Input**: token, new_password

**Steps**:
1) Hash token
2) Verify token row:
   - exists, not expired, not revoked, not used
3) Update user password:
   - hash new password
   - update user
4) Invalidate all refresh tokens:
   - revoke all `refresh_tokens` for user
5) Mark reset token used:
   - `UPDATE password_reset_tokens SET used_at=now(), revoked_at=now() WHERE token_hash=$1`

**Response**:
- `204 No Content` or `200 OK`

---

## 3.8 Change password (authenticated): `POST /auth/change-password`
**Input**: old_password, new_password

**Steps**:
1) Verify old password
2) Hash new password and update
3) Revoke all refresh tokens for user
4) (Optional) revoke sessions too

---

# 4) Middleware / authorization (what’s needed)

## Access token verification
For each protected request:
1) Extract access token
2) Verify signature + `exp` + `iss/aud` (if used)
3) Load user info only when needed (don’t query DB on every request unless needed)
4) Set `ctx.UserID` in Go context

### Needed
- Access token should be short-lived (5–15 minutes typical)
- Use `kid` for signing key rotation if you can

### Not needed
- DB lookup per request if JWT contains required claims (like `sub`)

---

# 5) Security stuff checklist (production)

## Transport / headers
- HTTPS + HSTS
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY` (or CSP frame-ancestors)
- `Referrer-Policy: strict-origin-when-cross-origin`

## Cookies
- refresh cookie: `HttpOnly; Secure; SameSite=Lax`
- set `Path=/auth/refresh` and maybe narrow domain
- clear cookie on logout

## Rate limiting
At minimum:
- `/auth/login` per IP and per email
- `/auth/refresh`
- `/auth/password-reset/request`
- `/auth/register`

## Logging (audit)
- Log event type + user_id + ip + user-agent
- Never log raw passwords or raw tokens

## DB constraints / indexes
- `users(email)` unique
- `refresh_tokens(token_hash)` already PK
- `refresh_tokens(user_id, revoked_at)` index for revoking all
- consider TTL cleanup job for expired rows

---

# 6) How to map this to your architecture (handler-service-repo)

### Handler
- Parse JSON / read cookies
- Validate request shape
- Call `AuthService.Login/Refresh/...`
- Set cookies and status codes

### Service
- Business logic: rotation, revoke, token generation, password verify
- Uses repositories
- Uses standard `context.Context`

### Repository
- SQL only:
  - `GetUserByEmail`
  - `InsertRefreshTokenHash`
  - `RevokeRefreshToken`
  - `GetVerificationTokenByHash`
  - `MarkUsedToken`
  - etc.

---

# 7) One thing you must decide: access token format
Choose one:

1) **JWT access token** (most common, easiest with short TTL)
2) **Opaque access token** + DB lookup (more DB load; sometimes used)

If you tell me which one you’re using already, I’ll align the exact claims and middleware flow.

---

## Quick “schema upgrade” question (so I can give exact SQL next)
Right now your `refresh_tokens` / `verification_tokens` / `password_reset_tokens` inherit from `tokens` and **only** `refresh_tokens` has `user_id`.

**Question:** Do `verification_tokens` and `password_reset_tokens` currently also include `user_id` in your actual database?  
- If not, we need to add it, otherwise the “activate user” and “reset password” flows can’t be implemented cleanly.

Reply with:
1) PostgreSQL version (if you know)
2) Whether `verification_tokens` and `password_reset_tokens` have `user_id` in your real schema
3) Are you using JWT access tokens already? (yes/no)