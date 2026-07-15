# Euler title is fetched from projecteuler.net, not user-supplied

ADR-0017 established the [[Problem Adder]] as network-free: it built a [[Problem]] [[Exercise]]
"from a number and title the user supplies," in deliberate contrast to the AoC [[Exercise Adder]]'s
network fetch. That made `--title` a **required** CLI flag and made an empty title a hard
validation error in `NewProblemAdder`. We **reverse the "title the user supplies" part of that
decision**: the [[Problem Title]] is now *derived* — fetched from projecteuler.net by a new
[[Problem Title Fetcher]] — exactly as an AoC puzzle's title is derived from its page. The `--title`
flag is removed. This makes the two Kinds symmetric: a title is a fact about the exercise, never an
input to it.

The change is worth recording because it is a user-facing contract change (a previously required
flag disappears) and because the resulting failure behaviour is deliberately **asymmetric with
AoC**, which a future reader would otherwise read as an inconsistency.

## The split failure contract

The AoC Adder hard-fails on any page-fetch failure: without the authenticated fetch there is no
per-user input, and an AoC exercise is useless without its input. Euler is different — a [[Problem]]
is fully solvable from its number alone; the title is cosmetic metadata. So the [[Problem Title
Fetcher]] distinguishes two failure modes:

- **Fetch succeeded but the page has no problem title** → the problem number does not exist (a
  typo, or a number beyond the archive). This is a **hard error**: nothing is scaffolded, so a
  mistyped number never leaves an orphaned directory behind. projecteuler.net does not return a
  clean 404 for a bad number — it serves a 200 with no problem heading — so "no title found in a
  successfully fetched page" is the *only* reliable signal that the number is invalid.
- **The site could not be reached** (network down, timeout, non-200) → a **transient** failure. The
  Adder logs a warning, substitutes the placeholder title `"Untitled"`, and scaffolds anyway. The
  command surfaces the placeholder to the user so they can correct `info.json` later. Being offline
  should never block creating a solvable exercise.

## Consequences

- **`--title` is removed from `add euler`.** Scripts or habits passing `-t "..."` break. There is
  no offline override to supply a title by hand at add time; the placeholder path covers the
  offline case, and `info.json` is hand-editable afterward.
- **The title-required invariant moves.** `NewProblemAdder` no longer rejects an empty title from
  the caller; instead the title always arrives non-empty (a real title, or the `"Untitled"`
  sentinel) because the fetcher guarantees it on the transient path and hard-errors on the
  bad-number path.
- **A new standalone, uncached fetcher.** The [[Problem Title Fetcher]] is separate from the AoC
  [[Page Fetcher]] rather than a refactor of it: the two sites differ in auth (Euler needs no
  token), key shape (bare number vs. year/day), and caching (none vs. on-disk). It reuses elf's
  `User-Agent` per site etiquette and parses the title with its own DOM walk, so drift in
  projecteuler.net's markup touches only the Euler path. It does not cache: `add euler` is a
  one-shot-per-problem operation, so a cache would buy almost nothing.
- **The placeholder is signalled, not silent.** The Adder returns whether the title was fetched or
  placeholdered; `cmd/add` prints a warning line after the scaffold report. The [[Scaffold
  Outcome]] Report stays purely a per-file record — the exercise-level warning travels beside it,
  not inside it.

## Considered options

**Keep `--title` as an offline override** (fetch by default, flag wins when present): rejected in
favour of removing the flag outright, for full symmetry with AoC where the title is never a user
input. The offline case is already covered by the placeholder path, so the flag earned its keep
only as a convenience, at the cost of a second, rarely-exercised code path and a muddier contract.

**Placeholder on every failure, including bad numbers**: rejected — it would silently scaffold an
`"Untitled"` exercise for a mistyped number, leaving a directory the user must notice and delete.
Distinguishing "bad input" (hard error) from "environment failure" (degrade) costs one branch and
makes typos fail loudly while keeping the offline experience graceful.

**Extend the AoC [[Page Fetcher]] to serve both sites**: rejected — the token requirement, cache
layout, and URL shape are all AoC-specific; parameterizing them would complicate the working AoC
path to save a small amount of duplicated resty/User-Agent setup.
