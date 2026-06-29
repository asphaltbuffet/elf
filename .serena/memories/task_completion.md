# Task Completion

Run before committing any phase:
1. `mise run lint` — must pass clean
2. `mise run test` — must pass clean

Full verification (preferred):
- `mise run dev` — generate + mock + lint + test + snapshot

Do NOT use bare `go test` or `go build` as the sole check — nix build and lint can catch things they miss.
