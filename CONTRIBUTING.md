# Contributing to go-p2p-escrow

Thank you for your interest in contributing!

## Developer Certificate of Origin (DCO)

All contributions must be signed off with the [Developer Certificate of Origin](https://developercertificate.org/). Add a `Signed-off-by` line to your commit messages:

```
git commit -s -m "your commit message"
```

## Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Write tests for your changes
4. Ensure all tests pass: `go test ./...`
5. Ensure no vet warnings: `go vet ./...`
6. Submit a pull request

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- All exported types and functions must have GoDoc comments
- Keep the public API surface minimal — when in doubt, keep it unexported
- Zero external dependencies in the core package is a hard constraint

## Adding a New Chain Adapter

1. Create `adapters/yourchain/` directory
2. Implement the `escrow.Escrow` interface
3. Add tests with at least 80% coverage
4. Add an example in `examples/yourchain-escrow/`
5. Update README.md's Supported Chains table

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include Go version, OS, and minimal reproduction steps
