# Auth

The auth service handles everything about logging in: registration, signin, Google OAuth,
email verification, password reset, and the whole JWT lifecycle. It is a Go gRPC service.

Path: `services/auth/` · Port: 50051 · Depends on: user service, Redis, SMTP, Google OAuth.

## What it does

- signs up new users, holds the pending signup in Redis, and sends a verification email
- completes registration once the email is verified, creating the account through the user service
- signs users in with a password and returns an access and a refresh token
- handles Google OAuth login and registration
- verifies and refreshes tokens, rotating the access and refresh pair
- serves its public keys as JWKS so the gateway can verify tokens
- runs forgot and reset password by email

## JWTs and key rotation

Tokens are signed RS256. The service keeps an RSA keystore and rotates the signing key on
a schedule, keeping older public keys around long enough that tokens signed with them
still verify. The gateway pulls these public keys through `GetPublicKeys` and verifies
every protected request against them.

## Email

A small worker inside the service consumes a Redis backed queue and sends the verification
and password reset emails over SMTP using HTML templates. Putting the send behind a queue
keeps the signup path fast and lets a slow mail server retry without blocking anyone.

## Endpoints

See [endpoints.md](endpoints.md) for the full `AuthService` RPC list.
