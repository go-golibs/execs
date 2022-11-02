package execs

import (
	"git.eth4.dev/golibs/errors"
)

// ExitEvent - событие завершения процесса раннера
type ExitEvent struct {
	Member Member
	Err    error
}

// ErrorTrace трассировка ошибок завершений процессов
type ErrorTrace []ExitEvent

func (trace ErrorTrace) Error() string {
	msg := "Exit trace for group:\n"

	for _, exit := range trace {
		if exit.Err != nil {
			msg += errors.Formatted(exit.Err).Error()
		}
	}

	return msg
}

func (trace ErrorTrace) ErrorOrNil() error {
	for _, exit := range trace {
		if exit.Err != nil {
			return trace
		}
	}

	return nil
}
