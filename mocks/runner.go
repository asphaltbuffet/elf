package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

var _ runners.Runner = new(MockRunner)

type MockRunner struct {
	mock.Mock
}

func (m *MockRunner) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRunner) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRunner) Run(task *runners.Task) (*runners.Result, error) {
	args := m.Called(task)
	return args.Get(0).(*runners.Result), args.Error(1)
}

func (m *MockRunner) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRunner) String() string {
	args := m.Called()
	return args.String(0)
}
