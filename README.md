# osto-auth-cli

A containerized, interactive REPL-style authentication CLI written in Go. Supports user registration, password-based login, optional TOTP two-factor authentication (Google Authenticator compatible), and session management — all backed by an embedded SQLite database and fully containerized with Docker Compose.

---

## Table of Contents

- [How It Works](#how-it-works)
- [Security Properties](#security-properties)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration Reference](#configuration-reference)
- [Commands](#commands)
- [Demo Script](#demo-script)
- [Database Schema](#database-schema)
- [Project Structure](#project-structure)

---

## How It Works

When you run the container, you drop directly into a REPL loop — a single persistent process with a prompt you type commands into, similar to `psql`, `redis-cli`, or a Python shell. There is no HTTP server, no REST API, and no separate client binary. The prompt itself changes to reflect your state:

```
osto>                   ← not logged in
osto(arnav)>            ← logged in as "arnav"
```

The available command set also changes dynamically: pre-login commands (`register`, `login`, `help`, `exit`) are replaced by post-login commands (`whoami`, `enable-2fa`, `disable-2fa`, `logout`, `help`, `exit`) the moment you authenticate — and swapped back the moment you log out or your session expires.

---

## Security Properties

| Property | Implementation |
|---|---|
| Password storage | bcrypt (no plaintext ever written to disk) |
| Session tokens | 32 random bytes (crypto/rand), only their SHA-256 hash is stored |
| TOTP secrets | AES-256-GCM encrypted at rest; raw secret never persisted |
| Credential errors | Identical message for "no such user" and "wrong password" — no enumeration |
| Account lockout | Configurable threshold and duration; applies to both password and TOTP failures |
| Input masking | Passwords and TOTP codes are masked at the prompt and excluded from session history |
| Raw error leakage | Internal errors (e.g. crypto failures) are logged to stderr only; users always see a deliberate, clean message |

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) installed and running.

That's it. No Go toolchain required — the build happens inside Docker.

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/arnavmahajan630/osto-cli.git
cd osto-cli
```

### 2. Create your environment file

```bash
cp .env.example .env
```

### 3. Generate a 32-byte encryption key

The application uses AES-256-GCM to encrypt stored TOTP secrets. It requires a base64-encoded 32-byte key, which you can generate with:

```bash
openssl rand -base64 32
```

Open `.env` and paste the output as the value of `APP_ENCRYPTION_KEY`. The app will refuse to start if this value is missing or decodes to anything other than exactly 32 bytes.

### 4. Run the CLI

```bash
docker compose run --rm app
```

This command:
1. Builds the Go binary inside a multi-stage Docker build (no local Go toolchain needed).
2. Runs the `migrate` container first, which applies the database schema to the persistent volume and exits.
3. Once migrations succeed, starts the `app` container with stdin/tty attached, dropping you into the REPL.

> **Starting fresh:** to wipe all data and start from a clean database, run `docker compose down -v` before the command above. This deletes the named volume holding the SQLite file.

---

## Configuration Reference

All configuration is read from environment variables at startup. Required variables cause a clean, logged failure if missing — the REPL never starts with invalid config.

| Variable | Required | Default | Description |
|---|---|---|---|
| `DB_PATH` | yes | — | Path to the SQLite file inside the container. Set to `/data/osto.db` in the provided Compose file. |
| `SESSION_TIME` | yes | — | How long a session stays valid after the last activity. Uses Go duration syntax: `15m`, `1h`, `30s`. |
| `APP_ENCRYPTION_KEY` | yes | — | Base64-encoded AES-256 key (must decode to exactly 32 bytes). Never logged, even partially. |
| `LOG_LEVEL` | no | `info` | Diagnostic log verbosity to stderr. One of `debug`, `info`, `warn`, `error`. |
| `CLEANUP_INTERVAL` | no | `1m` | How often the background goroutine sweeps and removes expired sessions. |
| `LOCKOUT_THRESHOLD` | no | `5` | Number of failed credential attempts (password or TOTP, combined) before an account is locked. |
| `LOCKOUT_DURATION` | no | `15m` | How long an account stays locked after hitting the threshold. |

---

## Commands

### Before login

| Command | Description |
|---|---|
| `register` | Create a new account. Prompts for username, password (masked), password confirmation, name, and an optional birth date. |
| `login` | Authenticate with username and password. If 2FA is enabled on the account, a TOTP code is requested as a second step. Both the password and TOTP prompts allow up to 3 attempts before giving up gracefully. |
| `help` | List all available commands. `help <command>` shows detailed usage for that command. |
| `exit` / `quit` / `q` | Exit the REPL. If a session is active, it is revoked before quitting. |

### After login

| Command | Description |
|---|---|
| `whoami` | Display your account details: username, name, registration date, MFA status, last login time, and current session expiry. |
| `enable-2fa` | Enroll TOTP two-factor authentication. Displays an ASCII QR code and a manual base32 secret fallback. The secret is only saved after you successfully verify a code — a failed verification leaves your account unchanged. Revokes your current session to force a fresh login with 2FA active. |
| `disable-2fa` | Remove TOTP authentication from your account after verifying your current code. Revokes your current session. |
| `logout` | Revoke the current session and return to the pre-login prompt. |
| `help` | Same as above — lists only the commands available in the current auth state. |
| `exit` / `quit` / `q` | Revokes the session and exits. |

> **Aliases:** `?` is an alias for `help`. Pressing Ctrl+C once prints a reminder; a second Ctrl+C cleanly revokes any active session and exits.

---

## Demo Script

Follow this sequence to exercise every feature end-to-end. All commands are typed at the `osto>` prompt.

**1. Register an account**
```
osto> register
Username: reviewer
Password: ••••••••••••
Confirm Password: ••••••••••••
Name: Reviewer Name
Birth Date (YYYY-MM-DD) [optional]: 2026-06-18
[OK] Registration successful. You can now log in.
```

**2. Log in (no 2FA yet)**
```
osto> login
Username: reviewer
Password: ••••••••••••
[OK] Welcome back, Reviewer Name.
```

**3. Check your identity**
```
osto(reviewer)> whoami
┌─────────────────────────────────────┐
│  Username        reviewer           │
│  Name            Reviewer Name      │
│  Registered      2026-06-18         │
│  MFA             Disabled           │
│  Last Login      just now           │
│  Session Expires in 14m 58s         │
└─────────────────────────────────────┘
```

**4. Enable 2FA**
```
osto(reviewer)> enable-2fa
[INFO] Scan the QR code below with Google Authenticator or Authy.
[INFO] Can't scan? Enter this code manually: JBSWY3DPEHPK3PXP

  ██████████████  ██  ██  ██████████████
  ...
  (ASCII QR code)
  ...

Enter the 6-digit code from your app to confirm setup: ••••••
[OK] 2FA enabled successfully. Please log in again.
```

**5. Log back in — now with 2FA**
```
osto> login
Username: reviewer
Password: ••••••••••••
[INFO] This account requires two-factor authentication.
Enter 2FA code: ••••••
[OK] Welcome back, Reviewer Name.
```

**6. Disable 2FA**
```
osto(reviewer)> disable-2fa
Enter your current 2FA code to confirm: ••••••
[OK] 2FA disabled successfully. Please log in again.
```

**7. Log in again (no 2FA prompt this time) then exit**
```
osto> login
Username: reviewer
Password: ••••••••••••
[OK] Welcome back, Reviewer Name.

osto(reviewer)> logout
[OK] Logged out successfully.

osto> exit
```

---

## Database Schema

The SQLite database uses three tables, applied automatically by the migration runner on first start.

```sql
-- User accounts
CREATE TABLE users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT    NOT NULL UNIQUE,
    password_hash   TEXT    NOT NULL,          -- bcrypt
    name            TEXT,
    birth_date      TEXT,                      -- YYYY-MM-DD, nullable
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at   DATETIME,
    mfa_enabled     BOOLEAN  NOT NULL DEFAULT 0,
    mfa_secret_enc  TEXT,                      -- AES-256-GCM ciphertext, base64(nonce||ct), nullable
    failed_attempts INTEGER  NOT NULL DEFAULT 0,
    locked_until    DATETIME                   -- NULL = not locked
);

-- Active and historical sessions
CREATE TABLE sessions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT     NOT NULL UNIQUE,  -- SHA-256 of the raw token; raw token never stored
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at      DATETIME NOT NULL,
    last_active_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at      DATETIME                   -- NULL = still active
);

-- Login audit trail
CREATE TABLE login_attempts (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    username     TEXT    NOT NULL,
    succeeded    BOOLEAN NOT NULL,
    attempted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason       TEXT                          -- e.g. 'bad_password', 'bad_totp', 'locked'
);
```

---

## Project Structure

```
osto-auth-cli/
├── cmd/osto-auth-cli/
│   └── main.go              # entry point: config → logger → db → services → REPL
├── internal/
│   ├── config/              # env loading, fail-fast validation
│   ├── repl/                # readline loop, dynamic prompt, command dispatch, completer
│   ├── commands/            # one file per command (register, login, whoami, …)
│   ├── auth/                # AuthService: register, login, TOTP session creation
│   ├── session/             # SessionService, AuthGuard
│   ├── totp/                # secret generation, QR rendering, code validation
│   ├── secure/              # bcrypt, AES-GCM, session token generation + hashing
│   ├── validation/          # username, password, date rules — pure, no I/O
│   ├── repository/          # UserRepository and SessionRepository (SQLite-backed)
│   ├── models/              # User and Session structs
│   ├── db/                  # connection setup, embedded migration runner
│   └── logger/              # slog wrapper, level config
├── migrations/
│   └── 0001_init.sql        # full schema, embedded into the binary via go:embed
├── Dockerfile               # multi-stage: Go build → minimal runtime, no cgo
├── docker-compose.yml       # migrate service (one-shot) + app service (interactive tty)
├── .env.example             # all variables documented with example values
└── README.md
```

## Further Improvements
 
The following are production-grade improvements that were intentionally left out of scope for this implementation, either for simplicity or because they go beyond the assignment's requirements. They represent a natural next step for hardening the system.
 
### Argon2id instead of bcrypt
 
The current implementation uses **bcrypt** for password hashing, which is a well-established and secure choice. However, **Argon2id** is the more modern recommendation — it won the Password Hashing Competition (2015) and is explicitly preferred by OWASP's current guidelines.
 
The practical advantage over bcrypt is that Argon2id is configurable across three dimensions: time cost (iterations), memory cost (KB of RAM used during hashing), and parallelism (threads). This makes it resistant to both GPU-based brute force (via the memory cost) and side-channel timing attacks (via the time cost), whereas bcrypt's hardness is controlled only by a single cost factor.
 
Switching would require: replacing `golang.org/x/crypto/bcrypt` with `golang.org/x/crypto/argon2`, storing the Argon2id parameters alongside the hash (time, memory, threads, salt) as a structured string in `password_hash`, and writing a migration path for existing bcrypt hashes — e.g. re-hash on next successful login transparently, without requiring a password reset.
 
### TOTP Backup Codes
 
If a user loses their authenticator app, there is currently no recovery path — the account becomes permanently inaccessible. Standard implementations (GitHub, Google) pair TOTP enrollment with a set of one-time backup codes the user saves offline. Each code is single-use and stored as individual hashed rows in a new `mfa_backup_codes` table. A `use-backup-code` command (or a branch in the login flow when the user answers "I don't have my device") consumes one and marks it used. The codes are generated and displayed only once, at enrollment time, alongside the QR code.
 
### Encryption Key Rotation
 
The `APP_ENCRYPTION_KEY` is currently a single static value. If it needs to be rotated (e.g. after a suspected infrastructure compromise), there is no in-place upgrade path — all stored `mfa_secret_enc` values would become undecryptable with the new key. A proper key rotation command (`rotate-key --old-key ... --new-key ...`) would decrypt every stored secret with the old key and re-encrypt it with the new one in a single transaction, making rotation a zero-downtime admin operation.
 
### Session Token Rotation on Each Request
 
Currently a session token is issued once at login and remains valid until it expires or is revoked. A more defensive pattern is **sliding-window rotation**: on each authenticated command, the old token is invalidated and a new one is issued. This limits the blast radius of a token being observed (e.g. from a process list or memory dump) — the observation window is at most one command cycle rather than the full session lifetime.
 
### Structured Audit Log
 
The `login_attempts` table provides a basic record of authentication events. A more complete audit log would record all post-login actions (2FA changes, session revocations) with timestamps and reasons, surfaced via a dedicated `audit-log` command available to the authenticated user. For multi-user deployments this is the minimum expectation for accountability.
 
### `change-password` Command
 
The current implementation has no way to change a password without deleting and re-registering the account. A `change-password` command would prompt for the current password (to re-authenticate the action in-session), then a new password twice, and re-hash. If TOTP is enabled, it should also require the current TOTP code before accepting the change — following the same "verify before mutating" pattern used in `enable-2fa` and `disable-2fa`.

