# Euler analyze is exercise-scope only; no cross-problem year graph

ADR-0017 deferred `benchmark` and `analyze` for Project Euler and, in its Consequences, committed
to a cross-problem view: *"`euler/` behaves like a year for Analysis Year scope: `elf analyze
euler/` produces a cross-problem comparison, the Euler analogue of a year graph."* Landing the
deferred work, we **reverse that specific consequence.** Euler is analyzable only at **exercise
scope** — one [[Problem]] at a time, rendered as the consistency-facet grid. There is no
cross-problem (year-scope) Euler graph.

The reason the AoC analogy breaks: an AoC year is a *curated set of ~25 puzzles of comparable,
escalating difficulty released together*, so plotting them on one axis tells a coherent "how did
each language do across this season" story. Euler problems are *independent, wildly heterogeneous
in difficulty, and sparsely selected by the user* (they might have solved 1, 2, 7, and 391). A
single cross-problem line graph over that set answers no question worth asking — the per-problem
language comparison is the only useful Euler analysis, and that is exactly what exercise scope
already gives.

## Consequences

- **Year scope becomes AoC-only.** `analyze` proceeds to a year graph only when the target holds at
  least one child and *every* child is a [[Puzzle]]. A Euler tree (children are Problems), a
  **mixed** tree (any Problem child), and an **empty** tree are all refused. The predicate is
  name-independent — it keys on the children's `Kind` (from their `info.json`), not on the
  directory being the configured `euler.dir` — so a symlinked, copied, or renamed Euler tree is
  still recognized.
- **`benchmark` and single-Problem `analyze` are now supported**, ending ADR-0017's deferral. The
  benchmarker iterates the exercise's *declared parts* (like the solver/tester already do), so a
  Problem runs only Part One and its `ImplementationData` has `PartTwo == nil`. The facet grid is
  data-present-driven, so a Problem renders a 1×N grid with no empty Part Two row.
- The refusal is surfaced through a single general `ErrUnsupportedAnalysis` sentinel (renamed from
  ADR-0017's `ErrEulerUnsupported`), with the specific case carried in the wrapped message rather
  than in a proliferation of one-off error vars. `ErrEulerUnsupported` — which previously guarded
  *both* benchmark and single-Problem analyze — is retired; benchmark no longer guards at all, and
  analyze guards only the year-scope-over-non-Puzzles case.
- A Problem's persisted `benchmark.json` records a dedicated `Number` field (Year/Day stay zero);
  the file's `String()` renders the Euler identity rather than a misleading `AOC 0/00`. The field
  is written but currently unread — the facet grid needs only the per-part sample data — kept as an
  honest on-disk record for future tooling.

## Considered options

**Build the cross-problem graph as ADR-0017 described**: rejected — the aggregate view is
uninformative for a sparse, heterogeneous problem set, so it would be effort spent producing a
misleading artifact.

**Let `analyze euler/` fall through to the year-graph path** (ScopeYear already classifies it as a
year): rejected — it renders exactly the meaningless line graph this ADR argues against. The
year-scope guard exists to prevent that fall-through.

**A scope-specific error name** (e.g. `ErrEulerCrossProblem`): rejected in favour of the general
`ErrUnsupportedAnalysis` — the sole user is the author, who will recognize the case from the
message, and a general sentinel guards against accreting a taxonomy of one-off unsupported-scope
errors.
