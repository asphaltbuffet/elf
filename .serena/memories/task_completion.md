# Task Completion

Run after any coding change:

```
mise run dev   # generate → mock → lint → test → snapshot
```

For quick iteration (skip snapshot):
```
mise run lint-fix && mise run test
```

After adding new files:
```
jj file track <path>   # required before nix build sees them
```

After dependency changes:
```
mise run mod-tidy   # auto-runs nix-hash (gomod2nix generate) as post-dep
```
