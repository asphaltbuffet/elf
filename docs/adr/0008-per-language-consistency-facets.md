# Exercise-scope graph shows per-language consistency facets

The exercise-scope (single-day) analyze graph is a **grid of per-language facets** that
compares **consistency** — how tightly each language's repeated runs cluster — rather than
speed. The grid is 2×N: rows are Parts (Part One, Part Two), columns are languages. Each
cell is an independently auto-scaled box plot whose Y axis is each sample's **percentage of
that language's own median** (sample ÷ median × 100), so the axis is dimensionless and
centred at 100%. The median (robust to outliers) is the normaliser; the absolute median is
shown in the cell's subtitle for context but is not the comparison axis. A missing
(language, part) renders as a blank-but-present cell so the grid stays aligned. This
**supersedes [ADR 0007](0007-relative-runtime-box-plot.md)** and replaces the relative box
plot outright.

The motivating evidence: ADR 0007 assumed a single shared axis could show both
cross-language speed *and* within-language consistency. Rendered against real data, it could
not. Two facts collide on one axis — the speed gap *between* languages spans ~4 orders of
magnitude (Go ≈ 1×, Python ≈ 30×, Bash ≈ 12,000× on 2015 day 1), which forces a log scale,
while the spread *within* a language is small (Bash ≈ ±5%, Go ≈ ±300% driven by a few
microsecond-scale outliers). A log axis sized for the between-language range crushes the
within-language spread to an invisible sliver, so the relative box plot rendered nearly
identically to the absolute one it replaced. Two throwaway prototypes confirmed the
tradeoff is real and unavoidable on one axis: the relative box plot shows speed but not
consistency; a per-language facet view (each cell auto-scaled to its own median) shows
consistency clearly but drops the cross-language speed axis into subtitle text. Since the
year-scope line graph already answers cross-language speed across days, the single-day view
is dedicated to the question the year graph cannot answer: per-language consistency.

## Considered options

**Keep the relative box plot (ADR 0007)**: rejected — proven illegible for consistency on
high-dynamic-range exercises, which is the common case (compiled vs interpreted vs shell).

**One combined image (relative box plot + facets stacked)**: keep speed *and* add
consistency in one PNG. Rejected for now — more layout complexity, and speed is already
covered by the year graph, so the single-day view does not need to re-answer it. Could be
revisited if a single-day speed view is later wanted.

**Raw absolute units on each facet axis** (instead of % of own median): rejected — panels
would not be shape-comparable across languages, and microsecond-scale jitter (Go) would look
alarming without the centred-at-100% framing that makes "±5% vs ±300%" directly readable.

**Box + all 10 sample points, or a pure dot/strip plot per cell**: considered because each
(language, part) has only ~10 samples, where showing raw points can be more honest than a
summary. Rejected in favour of a plain box plot to keep the grid uncluttered and visually
consistent with the existing box styling; gonum still surfaces true outliers as glyphs, so
spikes (e.g. Go's) remain visible.

**Mean as the normaliser** (instead of median): rejected — the mean is dragged by the
outliers that some languages clearly have, which would shift the whole box off-centre; the
median keeps each cell centred at 100% regardless of outliers.

## Consequences

- The exercise-scope graph loses the cross-language *speed* comparison from its axis
  entirely. That comparison is intentionally delegated to the year-scope absolute line
  graph, which is unchanged.
- The "color encodes language" invariant still holds (each facet's box is styled in the
  language colour), but cross-language *position* comparison is no longer meaningful — each
  cell has its own axis. The grid's value is comparing box *shapes* (spread), not heights.
- Rendering moves from a single plot to a tiled grid of plots; each cell needs its own
  median computation and auto-scaled axis. The relative-runtime machinery from ADR 0007
  (`fastestMeans`, `relativeSamples`, `RelativeLogTicks`, the 1× reference line) is no longer
  used by the exercise-scope graph and is removed unless reused elsewhere.
- `CONTEXT.md` *Comparison vocabulary* now defines **consistency** with the "% of own median"
  encoding; the exercise-scope description is updated from "relative box plot" to
  "consistency facets".
- Grid width grows with language count; acceptable because a single exercise typically has
  2–4 languages.
