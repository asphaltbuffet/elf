# Exercise-scope graph compares relative runtime, not absolute

The exercise-scope (single-day) analyze graph plots **relative runtime** — each language's
raw timing samples divided by the fastest language's mean *for that part* — on a log axis
with a reference line at 1×, instead of absolute durations. The box-plot shape is kept
(median, quartiles, outliers stay visible), but in ratio terms, so the comparison reads as
"how far behind the leader each language is, and how consistently." The previous absolute
box plot is replaced outright, not retained.

The motivating problem: on a single puzzle, languages span ~4 orders of magnitude (a real
2015 day-1 case: Go 1×, Python 30×, Bash 12,518×). On an *absolute* log axis every language
collapses into a flat band and the per-language comparison is unreadable — the exact "not
very informative" complaint that prompted this. Re-expressing the data as a ratio to the
per-part fastest anchors the reference at a fixed 1× line (log 0); a language's distance
above that line is then a directly legible "orders of magnitude behind the leader." A
language's *strength* shows as its box dropping toward — or reaching — 1×. Because it is a
same-day, same-machine ratio, it is inherently machine-independent and does not involve the
existing machine-calibration **normalization factor**, which is an unrelated concept (see
`CONTEXT.md`, *Comparison vocabulary*).

## Considered options

**Fixed reference language (e.g. always Go, or the configured `language`)**: anchor 1× to one
language across every graph. Rejected as the default in favour of *fastest-per-exercise* because
fastest-as-reference needs zero configuration and answers the most universal question ("how far
behind the winner"). Anchoring to the user's configured `language` is recorded as a possible
future override, not built now (YAGNI).

**Global (cross-part) baseline**: divide every sample by a single fastest mean across both parts.
Rejected — it distorts whichever part is slower overall and hides the case where the leader differs
between Part One and Part Two. The baseline is computed *per part*, so the language sitting at 1×
may differ between the two part-groups; that change is itself a strength signal.

**Keep both absolute and relative** (two files, a flag, or a two-panel image): rejected. The
absolute single-day box plot was judged uninformative, so retaining it adds output/config surface
(exercise scope writes exactly one graph) for a view the user does not want. Absolute magnitudes
remain available in `benchmark.json` and in the year-scope absolute line graph.

**Non-box forms (dot/strip plot, slopegraph)**: considered for showing ratio + spread + the
two parts. Rejected because the box shape already conveys the sample distribution, and the only
real defect was the absolute axis — fixing the axis reuses the existing rendering machinery rather
than introducing a new graph type.

## Consequences

- The exercise-scope reference line moves from the 15-second AoC soft-limit (absolute) to 1×
  (the per-part fastest). The 15s line is meaningless on a ratio axis and is dropped from this graph.
- The change is surgical: `collectBoxSamples`, the gonum `BoxPlot`, and the Part One/Part Two
  grouping are reused. The substantive additions are a per-part transform dividing samples by the
  fastest mean, an axis relabel to relative runtime, and the relocated reference line.
- A new comparison vocabulary is introduced in `CONTEXT.md` to keep **normalization factor**
  (machine calibration) distinct from **reference language** / **relative runtime** (cross-language
  comparison) and **consistency** (sample spread).
- Scope is single-exercise only. An across-days relative view (year scope) is explicitly out of
  scope here; the existing absolute year line graph is unchanged.
