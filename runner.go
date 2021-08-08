package execs

import (
	"os"
)

// RunFunc - раннер
type RunFunc func(signals <-chan os.Signal, ready chan<- struct{}) error

// Run - хэндлер раннера
func (r RunFunc) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	return r(signals, ready)
}
