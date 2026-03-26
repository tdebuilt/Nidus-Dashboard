# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
| < latest | No       |

Only the latest release receives security updates. We recommend always running the most recent version.

## Reporting a vulnerability

If you discover a security issue, **please do not open a public issue**.

Instead, report it privately:

1. **Email**: Send details to the maintainer (see GitHub profile for contact)
2. **GitHub Security Advisories**: Use the [private vulnerability reporting](https://github.com/tdebuilt/Nidus-Dashboard/security/advisories/new) feature

Please include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and aim to provide a fix or mitigation within 7 days for critical issues.

## Scope

The following are in scope:
- Authentication bypass or credential exposure
- SQL injection or command injection
- Cross-site scripting (XSS) in the web UI
- Unauthorized access to API endpoints
- Encryption weaknesses in config export/import

## Responsible disclosure

We kindly ask that you:
- Give us reasonable time to fix the issue before public disclosure
- Do not access or modify other users' data
- Do not perform actions that could harm the service availability
