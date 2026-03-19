# Auth endpoints

The auth service exposes one gRPC service, `AuthService`, defined in `proto/auth.proto`.
It is internal: only the gateway calls it.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `Signup` | `SignupRequest` to `Empty` | validate, store a pending signup in Redis, enqueue a verification email |
| `VerifyEmail` | `VerificationRequest` to `TokenResponse` | complete registration through the user service, return tokens |
| `Signin` | `SigninRequest` to `TokenResponse` | password auth, return an access and a refresh token |
| `OAuthURL` | `OAuthURLRequest` to `OAuthURLResponse` | build the provider login url |
| `OAuthCallback` | `OAuthCallbackRequest` to `TokenResponse` | finish OAuth login or registration |
| `VerifyToken` | `VerificationRequest` to `VerificationResponse` | internal token check |
| `RefreshToken` | `RefreshTokenRequest` to `TokenResponse` | rotate the access and refresh pair |
| `GetPublicKeys` | `Empty` to `GetPublicKeysResponse` | JWKS, so the gateway can verify signatures |
| `ForgotPassword` | `ForgotPasswordRequest` to `Empty` | send a reset email |
| `ResetPassword` | `ResetPasswordRequest` to `TokenResponse` | set a new password from a reset token |
