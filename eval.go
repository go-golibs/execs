package execs

import (
	"strings"
)

// EvalCmd - команда выполнения произвольного выражения в оболочке
type EvalCmd struct {
	args []string
}

// SessionName - имя запускаемой сессии оболочки
func (*EvalCmd) SessionName() string {
	return "eval"
}

// Args - аргументы команды
func (cmd *EvalCmd) Args() []string {
	return cmd.args
}

// Eval - вычислить произвольное выражение в оболочке
func Eval(expr string) (string, error) {
	if expr == "" {
		return "", nil
	}

	parts := strings.Split(expr, " ")
	command := &EvalCmd{}

	if len(parts) > 1 {
		command.args = parts[1:]
	}

	return Result(parts[0], command)
}
