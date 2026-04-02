# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 0.2.x | :x:               |

## Reporting a Vulnerability

If you discover a security vulnerability in Nidus Dashboard, please report it responsibly.

**Do NOT open a public issue.**

Instead, please send an email to **security@tdelab.eu** with:

- A description of the vulnerability
- Steps to reproduce the issue
- Any potential impact assessment

You should receive an acknowledgment within 48 hours. We will work with you to understand and address the issue before any public disclosure.

## Security Practices

Nidus Dashboard follows these security practices:

- **Credentials**: All service credentials are encrypted with AES-256-GCM before storage in the database
- **Authentication**: JWT-based authentication with optional TOTP 2FA
- **Password hashing**: bcrypt with appropriate cost factor
- **Rate limiting**: Applied to authentication endpoints
- **SSRF protection**: URL validation prevents access to private/internal IPs
- **Non-root container**: Docker image runs as unprivileged user (UID 1000)
