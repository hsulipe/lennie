# Lennie — Login Service

A Go authentication microservice that supports password-based login and federated identity via OIDC (Google OAuth). Includes a local Dex mock so you can develop and test the full OAuth flow without a real Google client.

---

## Architecture

The service follows a hexagonal (ports and adapters) layout:

```
cmd/
  main.go               — wires everything together, starts HTTP server
internal/
  domain/               — User, UserIdentity models
  ports/                — UserRepository interface
  services/             — business logic (SignIn, SignUp, OIDC)
  adapters/
    httphandler/        — HTTP handlers and router
    postgres/           — GORM/PostgreSQL repository implementation
```

Each `User` can have multiple `UserIdentity` rows — one per authentication provider (`password`, `google`, etc.). This lets a single account be reached via password or OAuth without duplicating user data.

---

## Services

| Service | Port | Description |
|---|---|---|
| `login` | `8080` | Authentication API |
| `db` | `5432` | PostgreSQL 15 |
| `dex` | `8081` | Local OIDC / Google OAuth mock (Dex) |
| `migration` | — | Flyway runs schema migrations on startup |

---

## Running locally

```bash
docker compose up --build
```

The stack is ready when you see:

```
login-1 | Server is running on http://localhost:8080
```

---

## API reference

### Health check

```bash
curl http://localhost:8080/health
```

### Password sign-up

```bash
curl -s -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ada Lovelace",
    "email": "ada@example.com",
    "provider": "password",
    "provider_id": "ada@example.com",
    "credentials": "secret123",
    "birthdate": "1815-12-10"
  }' | jq
```

### Password sign-in

```bash
curl -s -X POST http://localhost:8080/signin \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "password",
    "provider_id": "ada@example.com",
    "credentials": "secret123"
  }' | jq
```

---

## Testing the OIDC / Google OAuth flow

Dex acts as a local stand-in for Google. It issues real OIDC-compliant JWTs, so the full token exchange runs exactly as it would in production.

### Step 1 — open the login URL in a browser

```
http://localhost:8080/auth/login
```

You will be redirected to Dex. Two sign-in options are available:

| Option | Behaviour |
|---|---|
| **Mock (Google)** | One-click, auto-approves — no credentials needed. Good for automated flows. |
| **Email & Password** | Enter one of the static test accounts below. Closer to the real Google login screen. |

**Test accounts** (password: `password`):

| Email | Role |
|---|---|
| `user@example.com` | Regular user |
| `admin@example.com` | Admin user |

### Step 2 — callback

After Dex approves the login it redirects to:

```
http://localhost:8080/auth/callback?code=...&state=...
```

The login service exchanges the code, verifies the ID token, creates (or finds) the user in the database, and returns the user object as JSON.

### Testing with curl (automated)

The OIDC flow involves browser redirects, but you can drive it end-to-end with curl by following redirects and capturing cookies:

```bash
# 1. Start the login flow — capture the redirect to Dex
curl -v -c cookies.txt -b cookies.txt \
  "http://localhost:8080/auth/login" 2>&1 | grep "Location:"

# 2. Follow to Dex and use the mock connector (auto-approve)
#    The Location header from step 1 gives you the Dex URL.
#    Append &connector_id=mock to skip the login screen.
curl -v -c cookies.txt -b cookies.txt -L \
  "<DEX_AUTH_URL>&connector_id=mock" 2>&1 | grep "Location:"

# 3. Follow the callback redirect back to the login service
curl -s -c cookies.txt -b cookies.txt -L \
  "http://localhost:8080/auth/callback?code=<CODE>&state=<STATE>" | jq
```

For integration tests, prefer driving a headless browser (e.g. Playwright) over raw curl — it handles the full redirect chain automatically.

---

## Dex OIDC endpoints

Useful when debugging tokens or configuring a client:

| Endpoint | URL |
|---|---|
| Discovery document | `http://localhost:8081/dex/.well-known/openid-configuration` |
| JWKS (public keys) | `http://localhost:8081/dex/keys` |
| Authorization | `http://localhost:8081/dex/auth` |
| Token | `http://localhost:8081/dex/token` |

---

## SSO integration

Because Dex issues standard OIDC JWTs, any service in your platform can validate tokens without calling the login service at all — they just need Dex's public keys.

### How it works

```
Browser → GET /auth/login (login service)
        → redirect to Dex authorization endpoint
        → user authenticates
        → Dex issues an ID token (JWT) + access token
        → callback to login service (/auth/callback)
        → login service verifies token, returns user

Other services receive the ID token (passed by the client in the
Authorization header) and validate it independently using Dex's JWKS.
```

### Validating tokens in another service

Any downstream service can verify the Dex ID token with these parameters:

| Parameter | Value |
|---|---|
| Issuer (`iss`) | `http://localhost:8081/dex` (prod: your Dex domain) |
| JWKS URI | `http://localhost:8081/dex/keys` |
| Audience (`aud`) | `login-app` (the client ID) |

**Go example** using `go-oidc`:

```go
provider, _ := oidc.NewProvider(ctx, "http://localhost:8081/dex")
verifier := provider.Verifier(&oidc.Config{ClientID: "login-app"})

// In your HTTP middleware:
rawToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
idToken, err := verifier.Verify(r.Context(), rawToken)
if err != nil {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
    return
}

var claims struct {
    Email   string `json:"email"`
    Subject string `json:"sub"`
}
idToken.Claims(&claims)
```

**For production**, replace `http://localhost:8081/dex` with your real Dex/Google issuer URL. The token validation code stays the same — only the issuer changes.

### Replacing Dex with real Google OAuth

Change two things in `docker-compose.yaml` on the `login` service:

```yaml
- OIDC_ISSUER=https://accounts.google.com
- OIDC_PROVIDER_URL=https://accounts.google.com   # same as issuer; no Docker split needed
- OIDC_CLIENT_ID=<your-google-client-id>
- OIDC_CLIENT_SECRET=<your-google-client-secret>
- OIDC_REDIRECT_URL=https://your-domain.com/auth/callback
```

Remove the `dex` service from `docker-compose.yaml`. No code changes required.

---

## Local OAuth client credentials

These are defined in `dex/config.yaml` and only exist in the local mock:

| Field | Value |
|---|---|
| Client ID | `login-app` |
| Client secret | `login-app-secret` |
| Redirect URI | `http://localhost:8080/auth/callback` |
