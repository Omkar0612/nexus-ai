# Contributing to NEXUS

Thank you for helping make NEXUS better!

## Quick Start

```bash
git clone https://github.com/Omkar0612/nexus-ai.git
cd nexus-ai
cp config/nexus.example.toml ~/.nexus/nexus.toml
go mod tidy
make build
./bin/nexus start
```

## Ways to Contribute

- **Bug reports** — Use the bug report template
- **New skills** — Build a skill in the `skills/` directory
- **New agents** — Add a worker in `workers/`
- **Use cases** — Share real workflows in `examples/`
- **Star the repo** — Helps others discover NEXUS

## Code Style

- Go: `gofmt` + `golangci-lint`
- Python: `black` + `ruff`
- Commits: conventional commits (`feat:`, `fix:`, `docs:`)

## License

By contributing, you agree your code is MIT licensed.
