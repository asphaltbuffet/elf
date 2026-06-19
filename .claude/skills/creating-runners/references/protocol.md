# Runner Wire Protocol

The contract every runner's wrapper template must implement. It is a **public contract**
(ADR-0002), not a runner internal — out-of-process runners speak it without importing Go code.

**Source of truth:** `pkg/protocol/` (Task, Result, Part). Read it if anything here looks stale.
Reference harnesses: `pkg/runners/interface/go.tmpl` and `python.templ`.

## Transport

Line-delimited JSON over the subprocess's stdin/stdout. One JSON object per line.

- elf writes one **Task** per line to the subprocess **stdin**.
- the harness writes one **Result** per line to **stdout**, and flushes.
- Any stdout line that does **not** parse as a Result is treated as debug output and printed by
  elf — it is not an error. So `print`/`eprintln` for debugging is safe; just don't emit a line
  that accidentally looks like a Result.

## Task (elf → harness)

```json
{"task_id": "string", "part": 1, "input": "string", "output_dir": "string"}
```

- `task_id` — opaque; echo it back unchanged in the matching Result.
- `part` — `1` = PartOne, `2` = PartTwo, `3` = Visualize. Frozen wire values.
- `input` — the puzzle input (or a test case's input).
- `output_dir` — present only for Visualize (part 3); where to write visual artifacts. May be absent.

## Result (harness → elf)

```json
{"task_id": "string", "ok": true, "output": "string", "duration": 0.0}
```

- `task_id` — the value from the Task being answered.
- `ok` — `true` if the solution ran; `false` on error/panic.
- `output` — the answer string (or the error message when `ok:false`).
- `duration` — wall-clock seconds the solution took (float).

## Harness responsibilities

1. Read stdin line by line until EOF; one Task per line.
2. Parse the Task; dispatch on `part` to the user's solution function.
3. Time the call; build a Result with `task_id`, `ok`, `output`, `duration`.
4. **Catch panics/exceptions** — emit `ok:false` with the message rather than crashing the
   subprocess. A crash loses the Result and hangs the run.
5. Write the Result as one JSON line; flush stdout.
6. Loop. Exit cleanly on EOF.

## Part dispatch

| part | meaning | user function (signature is your design — confirm during grilling) |
|------|---------|--------------------------------------------------------------------|
| 1 | PartOne | `part_one(input) -> answer` |
| 2 | PartTwo | `part_two(input) -> answer` |
| 3 | Visualize | `visualize(input, output_dir) -> answer` — writes artifacts to `output_dir` |

Only implement the parts the user wants (most AoC solutions are parts 1 and 2; Visualize is opt-in).
