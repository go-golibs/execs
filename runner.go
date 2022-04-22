package execs

import "os"

// Runner - стартер управляемого метода, объекта либо груупы таковых
type Runner interface {
	// Run хэндлер функции стартера
	Run(signals <-chan os.Signal, ready chan<- struct{}) error
}

// RunFunc - функция стартер
type RunFunc func(signals <-chan os.Signal, ready chan<- struct{}) error

// Run - хэндлер функции стартера
func (r RunFunc) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	return r(signals, ready)
}
