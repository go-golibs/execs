package execs

import (
	"os"
	"os/exec"
	"path/filepath"

	"git.corout.in/golibs/buffer"
)

// Command - инерфейс команды системной командной оболочки
type Command interface {
	Args() []string
	SessionName() string
}

// Enverer - интерфейс объектов использующих переменные окружения
type Enverer interface {
	Env() []string
}

// WorkingDirer - интерфейс объектов использующих рабочую директорию
type WorkingDirer interface {
	WorkingDir() string
}

// BufferProvider - буферезируемый объект
type BufferProvider interface {
	Buffer() *buffer.Buffer
}

// Exiter - завершаемый объект
type Exiter interface {
	ExitCode() int
}

// Runner - запускатор всякого
type Runner interface {
	Run(signals <-chan os.Signal, ready chan<- struct{}) error
}

// NewCommand - конструтор команды оболочки
// nolint
func NewCommand(path string, command Command) *exec.Cmd {
	cmd := exec.Command(filepath.Clean(path), command.Args()...)
	cmd.Env = os.Environ()

	if ce, ok := command.(Enverer); ok {
		cmd.Env = append(cmd.Env, ce.Env()...)
	}

	if wd, ok := command.(WorkingDirer); ok {
		cmd.Dir = wd.WorkingDir()
	}

	return cmd
}
