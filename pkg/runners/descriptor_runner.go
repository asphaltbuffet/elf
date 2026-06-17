package runners

import (
	"context"
	"errors"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// errNotImplemented is returned by stub lifecycle methods pending Task 4.
var errNotImplemented = errors.New("not implemented")

// descriptorRunner implements Runner using a RunnerDescriptor.
type descriptorRunner struct {
	desc RunnerDescriptor
	meta ExerciseMeta
}

func (r *descriptorRunner) String() string { return r.desc.Name }

// Prepare, Open, Run, Close, and Cleanup will be fully implemented in Task 4.

func (r *descriptorRunner) Prepare(_ context.Context) error { return nil }

func (r *descriptorRunner) Open(_ context.Context) error { return nil }

func (r *descriptorRunner) Run(_ context.Context, _ *protocol.Task) (*protocol.Result, error) {
	return nil, errNotImplemented
}

func (r *descriptorRunner) Close(_ context.Context) error { return nil }

func (r *descriptorRunner) Cleanup() error { return nil }
