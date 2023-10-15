// Package common contains the base struct for all exercises.
package common

import (
	"errors"
	"io"
)

// BaseExercise is the base struct for all exercises.
type BaseExercise struct{}

// One is the first part of the exercise.
//
//nolint:revive // this is a stub
func (e BaseExercise) One(in string) (any, error) {
	return nil, errors.New("not implemented")
}

// Two is the second part of the exercise.
//
//nolint:revive // this is a stub
func (e BaseExercise) Two(in string) (any, error) {
	return nil, errors.New("not implemented")
}

// Vis is the visualization of the exercise.
//
//nolint:revive // this is a stub
func (e BaseExercise) Vis(in string, w *io.Writer) error {
	return errors.New("not implemented")
}
