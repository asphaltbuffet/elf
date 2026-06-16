// Package protocol defines the wire types shared between elf and out-of-process runners.
// Task, Result, and Part constitute the public contract for the stdin/stdout JSON protocol.
package protocol

// Part identifies which sub-problem of a puzzle a Task addresses.
// These values are part of the wire protocol — never change them without
// updating all runner implementations.
type Part uint8

// Exercise part constants. Values are stable wire protocol identifiers.
const (
	PartOne   Part = 1
	PartTwo   Part = 2
	Visualize Part = 3
)

// Task is a unit of work sent to a Runner.
type Task struct {
	TaskID    string `json:"task_id"`
	Part      Part   `json:"part"`
	Input     string `json:"input"`
	OutputDir string `json:"output_dir,omitempty"`
}

// Result is the outcome of a Task.
type Result struct {
	TaskID   string  `json:"task_id"`
	Ok       bool    `json:"ok"`
	Output   string  `json:"output"`
	Duration float64 `json:"duration"`
}
