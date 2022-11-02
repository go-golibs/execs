package execs

import (
	"os"
	"os/exec"
	"path/filepath"

	"git.eth4.dev/golibs/iorw"
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

// BufferProvider - буферизируемый объект
type BufferProvider interface {
	Buffer() *iorw.Buffer
}

// Exiter - завершаемый объект
type Exiter interface {
	ExitCode() int
}

// NewCommand - конструктор команды оболочки
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
