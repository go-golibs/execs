package execs

import (
	"io"
	"os/exec"
	"strings"
	"sync"

	"git.corout.in/golibs/buffer"
	"git.corout.in/golibs/errors"
)

const (
	haveStdoutParam = 1
	haveStderrParam = 2
)

// Result - возвращает результат выполнения команды оболочки из stdout
// ctx обязан содержать zap.Logger
func Result(name string, command Command, writers ...io.Writer) (string, error) {
	sess, err := Start(NewCommand(name, command))
	if err != nil {
		return "", errors.Wrap(err, "run command session")
	}

	sess = sess.Wait()

	for e := range sess.errors {
		err = errors.And(err, e)
	}

	return strings.Trim(string(sess.Buffer().Contents()), "\n"), err
}

// Run - запускает команду оболочки
func Run(name string, command Command, writers ...io.Writer) error {
	sess, err := Start(NewCommand(name, command), writers...)
	if err != nil {
		return errors.Wrap(err, "run command session")
	}

	sess.Wait()

	return nil
}

// Start - запускает команду оболочки завернутую в управляющий враппер
func Start(cmd *exec.Cmd, writers ...io.Writer) (*Session, error) {
	exited := make(chan struct{})

	session := &Session{
		cmd:      cmd,
		bout:     buffer.NewBuffer(),
		berr:     buffer.NewBuffer(),
		errors:   make(chan error),
		exited:   exited,
		lock:     &sync.Mutex{},
		exitCode: -1,
	}

	var (
		wcount         = len(writers)
		cmdout, cmderr io.Writer
	)

	cmdout, cmderr = session.bout, session.berr

	if wcount >= haveStdoutParam {
		cmdout = io.MultiWriter(cmdout, writers[0])
	}

	if wcount >= haveStderrParam {
		cmderr = io.MultiWriter(cmderr, writers[1])
	}

	cmd.Stdout, cmd.Stderr = cmdout, cmderr

	err := cmd.Start()
	if err == nil {
		go session.monitorForExit(exited)
	}

	return session, err
}
