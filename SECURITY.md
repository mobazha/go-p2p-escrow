# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in go-p2p-escrow, **please do not open a public issue**.

Instead, report it via email:

**security@mobazha.org**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide a timeline for a fix.

## Scope

The following are in scope:
- State machine bypass (allowing invalid state transitions)
- Fund loss scenarios (signatures accepted when they shouldn't be)
- Key material leakage through API or logs
- Denial of service on the Registry or Store

## Supported Versions

| Version | Supported |
|---|---|
| latest (main branch) | ✅ |
| tagged releases | ✅ |

## Responsible Disclosure

We follow a 90-day responsible disclosure timeline. We request that you:
1. Allow us reasonable time to fix the issue before public disclosure
2. Make a good faith effort to avoid data destruction or service disruption
3. Do not access or modify other users' data
