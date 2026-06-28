# Task Completion

Run before committing any change:

```
mise run lint   # must pass clean
mise run test   # must pass with coverage
```

For a full verification (including generate, mock, snapshot):
```
mise run dev
```

For auto-fixable lint issues, run `mise run lint-fix` first, then `mise run lint` to confirm clean.
