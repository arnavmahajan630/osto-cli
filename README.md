# Osto Auth CLI

An interactive, REPL-style CLI for user registration, login, optional TOTP 2FA, and session management. It is backed by an embedded SQLite database and is fully containerized using Docker Compose.

## Prerequisites

- **Docker** and **Docker Compose** installed on your system.

## Environment Setup

1. Copy the provided `.env.example` file to create your local `.env` configuration.
   ```bash
   cp .env.example .env
   ```
2. Generate a valid 32-byte (256-bit) encryption key for AES-GCM encryption and add it to your `.env` file. You can easily generate one using OpenSSL:
   ```bash
   openssl rand -base64 32
   ```
   *Make sure `APP_ENCRYPTION_KEY` is set to this exact string in the `.env` file.*

## Running the Application

This CLI is fully containerized. A `migrate` container runs sequentially before the main `app` container to ensure your SQLite database has the proper schema. 

Since the application requires an interactive terminal (REPL), you must run it using Docker Compose with standard input attached.

### 1. Build and Run the CLI
Use the following command to completely delete any old anonymous containers, apply database migrations, and boot directly into the REPL loop:

```bash
docker compose run --rm app
```

*(Note: If you want to start completely fresh with a new database, you can run `docker compose down -v` to delete the local `osto_data` persistent volume before running the above command).*

## Step-by-Step Demo Script

You can follow this exact sequence to comprehensively test the features and security properties of the CLI.

**1. Register an Account**
```text
osto> register
Username: reviewer
Password: Password123
Name: Reviewer Name
Birth Date (YYYY-MM-DD) [Optional]: 2026-06-18
[OK] Registration successful
```

**2. Enable 2FA**
```text
osto> enable-2fa
```
*The terminal will render an ASCII QR code. Scan it with a TOTP app (like Google Authenticator or Authy).*
```text
Enter code: <type the 6-digit code from your app>
[OK] 2FA enabled successfully
```

**3. Logout**
```text
osto> logout
[OK] Logged out successfully
```

**4. Login (Now requiring 2FA)**
```text
osto> login
Username: reviewer
Password: Password123
[INFO] 2FA is required for this account.
Enter 2FA Code: <type the current 6-digit code>
[OK] Login successful
```

**5. Check Authenticated Identity**
```text
osto> whoami
[OK] Current User: reviewer (Name: Reviewer Name)
```

**6. Disable 2FA**
```text
osto> disable-2fa
Enter code: <type the current 6-digit code>
[OK] 2FA disabled successfully
```

**7. Logout & Exit**
```text
osto> logout
[OK] Logged out successfully
osto> exit
```
