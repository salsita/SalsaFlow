package common

import (
	// Internal
	"github.com/salsaflow/salsaflow/errs"
)

// ErrNotStageable shall be returned from EnsureStageable()
// when the given release cannot be staged yet.
//
// ErrNotStageable implements errs.Error interface.
type ErrNotStageable struct {
	// Cannot simply embed because then the name collides
	// with the Error method. Sucks.
	E errs.Error
}

func (err *ErrNotStageable) Error() string {
	return err.E.Error()
}

func (err *ErrNotStageable) Hint() string {
	return err.E.Hint()
}

func (err *ErrNotStageable) Task() string {
	return err.E.Task()
}

func (err *ErrNotStageable) Err() error {
	return err.E.Err()
}

type ErrNotReleasable struct {
	// Cannot simply embed because then the name collides
	// with the Error method. Sucks.
	E errs.Error
}

// ErrNotReleasable shall be returned from EnsureReleasable()
// when the given release cannot be released yet.
//
// ErrNotReleasable implements errs.Error interface.
func (err *ErrNotReleasable) Error() string {
	return err.E.Error()
}

func (err *ErrNotReleasable) Hint() string {
	return err.E.Hint()
}

func (err *ErrNotReleasable) Task() string {
	return err.E.Task()
}

func (err *ErrNotReleasable) Err() error {
	return err.E.Err()
}
