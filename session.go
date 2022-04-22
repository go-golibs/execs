package execs

import (
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"git.corout.in/golibs/errors"
	"git.corout.in/golibs/iorw"
)

const (
	// InvalidExitCode - код возврата процесса "выход с ошибкой"
	InvalidExitCode = 254

	// ExitCodePrefix - префикс кода возвращаемого процессом
	ExitCodePrefix = 128
)

// Session - враппер для запуска и управления процессом командной оболочки
type Session struct {
	cmd      *exec.Cmd
	bout     *iorw.Buffer
	berr     *iorw.Buffer
	exited   <-chan struct{}
	lock     *sync.Mutex
	errors   chan error
	exitCode int
}

// Buffer - возвращает буфер вывода процесса сессии
func (s *Session) Buffer() *iorw.Buffer {
	return s.bout
}

// ExitCode - возвращает код возврата процесса
func (s *Session) ExitCode() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.exitCode
}

// ErrLog - возвращает канал в который отправляются
// ошибки времени выполнения, можно использовать в логере
func (s *Session) ErrLog() <-chan error {
	return s.errors
}

// Kill - посылает системный сигнал `kill -9` процессу
func (s *Session) Kill() *Session {
	return s.Signal(syscall.SIGKILL)
}

// Interrupt - посылает системный сигнал `kill -2` процессу
func (s *Session) Interrupt() *Session {
	return s.Signal(syscall.SIGINT)
}

// Terminate - посылает системный сигнал `kill -15` процессу
func (s *Session) Terminate() *Session {
	return s.Signal(syscall.SIGTERM)
}

// Signal - посылает произвольный сигнал процессу
func (s *Session) Signal(signal os.Signal) *Session {
	if s.processIsAlive() {
		if err := s.cmd.Process.Signal(signal); err != nil {
			s.errors <- errors.Ctx().
				Stringer("signal", signal).
				Wrap(err, "send signal to proc")
		}
	}

	return s
}

// Wait - ожидает завершения процесса
func (s *Session) Wait(timeout ...interface{}) *Session {
	// если указан таймаут, обрабатываем
	if len(timeout) >= 1 {
		if t, ok := timeout[0].(time.Duration); ok {
			var terminated *Session

			timer := time.AfterFunc(t, func() {
				terminated = s.Terminate()
			})

			for {
				select {
				case <-timer.C:
					timer.Stop()

					return terminated
				case <-s.exited:
					return s
				}
			}
		}
	}

	<-s.exited

	return s
}

func (s *Session) monitorForExit(exited chan<- struct{}) {
	err := s.cmd.Wait()
	if err != nil {
		s.errors <- errors.Wrap(err, "command return")
	}

	s.lock.Lock()

	if err = s.bout.Close(); err != nil {
		s.errors <- errors.Wrap(err, "close output buffer")
	}

	if err = s.berr.Close(); err != nil {
		s.errors <- errors.Wrap(err, "close error buffer")
	}

	status, _ := s.cmd.ProcessState.Sys().(syscall.WaitStatus)

	if status.Signaled() {
		s.exitCode = ExitCodePrefix + int(status.Signal())
	} else {
		exitStatus := status.ExitStatus()
		if exitStatus == -1 && err != nil {
			s.exitCode = InvalidExitCode
		}
		s.exitCode = exitStatus
	}
	s.lock.Unlock()

	close(exited)
	close(s.errors)
}

func (s *Session) processIsAlive() bool {
	return s.ExitCode() == -1 && s.cmd.Process != nil
}
