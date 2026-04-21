# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | ✅         |

## Reporting a Vulnerability

If you discover a security vulnerability in `go-p2p-escrow`, please report it responsibly.

**Do NOT open a public GitHub issue.**

Instead, email **security@mobazha.org** with:

1. A description of the vulnerability
2. Steps to reproduce
3. The potential impact
4. Any suggested fix (optional)

We will acknowledge receipt within 48 hours and aim to provide a fix within 7 days for critical issues.

## Security Design Principles

- **Key zeroing**: Private key material is overwritten in memory after use
- **State machine enforcement**: All state transitions go through `StateMachine` to prevent fund loss
- **Sentinel errors**: Error messages do not expose internal state
- **No logging of secrets**: Private keys, mnemonics, and signing data are never logged

## Scope

This policy covers the `go-p2p-escrow` Go module. For vulnerabilities in the Mobazha platform, see [mobazha.org](https://mobazha.org).
