# Runner lifecycle split into explicit stages

The original `Runner` interface collapsed build, launch, and process lifecycle into `Start()`.
This made failure modes ambiguous and forced every runner to implement steps it didn't need
(e.g. a PATH-discovered runner has no build step).

The interface is split into: `Prepare` (compile/write artifacts — no-op for pre-built runners),
`Open` (start the subprocess), `Run` (send one task, receive one result), `Close` (graceful
stop), `Cleanup` (remove artifacts — no-op where not applicable). All stages except `Cleanup`
accept a `context.Context` for cancellation and timeout. `Open`/`Close` mirrors the standard
Go resource lifecycle idiom (`os.File`, `sql.DB`, `net.Conn`). This also fixes the hardcoded
`context.Background()` calls in the original implementation.
