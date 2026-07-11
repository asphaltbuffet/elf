# `encrypt` leaves plaintext in place; the `.age` file is a derived artifact

`elf encrypt <exercise>` writes a per-file `.age` ciphertext for each file in the
[[Solution Set]] (`info.json` plus every language subdirectory) so a [[Problem]] can be committed to
a public repository without publishing its answer, satisfying Project Euler's "do not share
solutions" rule. The obvious implementation — mirroring age's own CLI (`age -R … file > file.age`
followed by deleting `file`) — would **remove the plaintext** after sealing it, so that only
ciphertext remains on disk.

We decided the opposite: `encrypt` **never removes the plaintext**. The `.age` files are treated as
**derived artifacts** (like build output). The plaintext stays the source of truth on the working
machine; the user gitignores the plaintext and commits the `.age` files. `decrypt` exists to
reconstruct the plaintext on a *fresh clone* that carries only the committed ciphertext.

The reason is data-safety, not convenience. Because `encrypt` removes nothing, a crash, a bad
recipient list, or a killed process mid-run can never lose a solution — the worst case is a stale or
missing `.age`, always regenerable from the plaintext that is still there. The cost of remaining
safe is that git hygiene moves entirely to the user: elf does not guarantee the plaintext stays out
of a commit; a correct `.gitignore` does. We judged that an acceptable trade for making the tool
incapable of destroying the user's work.

## Considered options

- **Leave plaintext in place; `.age` is a committed artifact (chosen).** Zero data-loss risk;
  re-running `encrypt` refreshes every `.age` from the current plaintext (reporting
  [[Scaffold Outcome]] Added/Replaced). Git safety is the user's responsibility via `.gitignore`.
- **Remove plaintext after sealing (age-CLI style).** Automatic secret hygiene — plaintext can never
  be committed because it does not exist post-encrypt. Rejected: it makes elf a custodian that can
  lose data on any partial failure, and it cannot re-encrypt an *edited* solution (nothing plaintext
  is left to re-read).
- **Move plaintext to a gitignored scratch location.** A middle path, but it still relocates the
  user's authoritative files and complicates every other verb, which expects solutions at their
  canonical path.

## Consequences

- `encrypt` is idempotent as a *refresh*: an existing `.age` whose plaintext changed is **Replaced**,
  not skipped, so ciphertext never goes stale relative to its source.
- `decrypt` does not clobber by default: if plaintext already exists it is **Skipped** (`--force`
  makes it **Replaced**), because on a working machine the plaintext — not the ciphertext — is
  authoritative.
- Users must gitignore the plaintext ([[Solution Set]] members) and commit the `.age` siblings. This
  is documented in the command help; elf enforces none of it.
