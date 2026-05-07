# Contributing to openbattery

Thanks for your interest! 🎉

## Reporting Bugs

1. Check if the issue already exists in [Issues](../../issues).
2. If not, open a new issue using the **Bug report** template.
3. Provide as much detail as possible:
   - macOS version
   - openbattery version (or commit hash)
   - what happened vs. what you expected
   - steps to reproduce
   - screenshots (if UI-related)

## Suggesting Features

1. Open an issue using the **Feature request** template.
2. Describe:
   - what you want to achieve
   - why it's needed
   - any implementation ideas

## Pull Requests

1. **Fork** the repo and clone it locally.
2. Make sure you have installed:
   - [Go](https://go.dev/dl/) ≥ 1.21
   - `golangci-lint` (optional but recommended)
3. Run tests and linters before committing:
   ```bash
   go vet ./...
   go test -race -v ./...
   golangci-lint run
   ```
4. Create a branch with a descriptive name:
   ```bash
   git checkout -b fix/short-description
   # or
   git checkout -b feat/short-description
   ```
5. Make your changes. Please:
   - follow the existing code style
   - add tests where possible
   - avoid leaving `// TODO` comments without explanation
6. Commit with descriptive messages:
   ```
   fix: correct percentage rounding in dashboard
   feat: add dark mode toggle
   ```
7. Open a Pull Request and fill out the template.

## Running Locally

```bash
go mod download
go build -o openbattery .
./openbattery
```

## Tests

The project uses standard `go test`:

```bash
go test ./...
```

If you add a new module, please try to write tests for it.

## Style Guide

- Formatting: `gofmt` (or `goimports`).
- Names in English.
- Comments on exported functions/types following Go conventions.
- Do not commit the `openbattery` binary — it's in `.gitignore`.

## Questions?

Open a [Discussion](../../discussions) or comment in an issue — happy to help!
