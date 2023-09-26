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
	return m.Called().Error(0)
}

func (m *MockRunner) Cleanup() error {
	return m.Called().Error(0)
}

func (m *MockRunner) Run(task *runners.Task) (*runners.Result, error) {
	return m.Called(task).Get(0).(*runners.Result), m.Called(task).Error(1)
}

func (m *MockRunner) Stop() error {
	return m.Called().Error(0)
}
