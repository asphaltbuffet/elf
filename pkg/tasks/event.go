package tasks

import "github.com/asphaltbuffet/elf/pkg/protocol"

// EventKind identifies a stage in a task's lifecycle as it streams to a renderer.
type EventKind int

// Task lifecycle event kinds. Planned/Started/Finished are emitted per task.
const (
	eventInvalid  EventKind = iota // zero value guards against an unset Kind
	EventPlanned                   // task will run (renderer shows "<not started>")
	EventStarted                   // task is now executing (renderer animates spinner + timer)
	EventFinished                  // task is done; Result is non-nil
)

// Event is a presentation-facing record of a task's lifecycle. It is display-only:
// the authoritative outcome of a run remains the []Result returned by the operation.
// Result is non-nil only when Kind == EventFinished.
type Event struct {
	Kind    EventKind
	ID      string
	Type    TaskType
	Part    protocol.Part
	SubPart int
	Result  *Result
}

// PlannedEvent reports that a task will run.
func PlannedEvent(id string, t TaskType, part protocol.Part, subPart int) Event {
	return Event{Kind: EventPlanned, ID: id, Type: t, Part: part, SubPart: subPart}
}

// StartedEvent reports that a task has begun executing.
func StartedEvent(id string, t TaskType, part protocol.Part, subPart int) Event {
	return Event{Kind: EventStarted, ID: id, Type: t, Part: part, SubPart: subPart}
}

// FinishedEvent reports that a task has completed, carrying its [Result].
func FinishedEvent(r Result) Event {
	return Event{
		Kind:    EventFinished,
		ID:      r.ID,
		Type:    r.Type,
		Part:    r.Part,
		SubPart: r.SubPart,
		Result:  &r,
	}
}
