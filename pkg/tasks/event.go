package tasks

import "github.com/asphaltbuffet/elf/pkg/protocol"

// EventKind identifies a stage in a task's lifecycle as it streams to a renderer.
type EventKind int

// Task lifecycle event kinds. Meta is emitted once at the start of a run;
// Planned/Started/Finished are emitted per task.
const (
	eventInvalid  EventKind = iota // zero value guards against an unset Kind
	EventMeta                      // run metadata for chrome (header, pretty language); Meta is non-nil
	EventPlanned                   // task will run (renderer shows "<not started>")
	EventStarted                   // task is now executing (renderer animates spinner + timer)
	EventFinished                  // task is done; Result is non-nil
)

// Meta carries the run-level metadata a renderer needs to draw chrome: the
// exercise identity for the header box and the runner's human-readable name for
// section labels. It is resolved by the domain after the exercise is loaded.
type Meta struct {
	Year     int
	Day      int
	Title    string
	Language string // human-readable runner name (e.g. "Rust"), not the lookup key ("rs")
}

// Event is a presentation-facing record of a run's progress. It is display-only:
// the authoritative outcome of a run remains the []Result returned by the operation.
// Result is non-nil only when Kind == EventFinished; Meta is non-nil only when
// Kind == EventMeta.
type Event struct {
	Kind    EventKind
	ID      string
	Type    TaskType
	Part    protocol.Part
	SubPart int
	Result  *Result
	Meta    *Meta
}

// MetaEvent reports run-level chrome metadata. It should be emitted once, before
// any task events, so a renderer can draw the header and section labels.
func MetaEvent(m Meta) Event {
	return Event{Kind: EventMeta, Meta: &m}
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
