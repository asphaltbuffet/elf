# Runner protocol as a public contract

The wire types (Task, Result) and the stdin/stdout JSON protocol are a public contract, not
runner internals. They live in a dedicated package (`pkg/protocol` or similar) so that
out-of-process runners — discovered at runtime via PATH or a plugin directory — can implement
the same protocol without importing runner implementation code. Built-in runners (Go, Python)
are reference implementations. This enables adding language support without recompiling elf,
analogous to the Terraform provider model.
