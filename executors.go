package execs

import (
	"bytes"
	"io"
	"os/exec"
	"sync"

	"git.corout.in/golibs/errors"
	"git.corout.in/golibs/iorw"
)

const (
	haveStdoutParam = 1
	haveStderrParam = 2
)

// Result - возвращает результат выполнения команды оболочки из stdout
// ctx обязан содержать zap.Logger
func Result(name string, command Command, writers ...io.Writer) (string, error) {
	sess, err := StartCmd(NewCommand(name, command), writers...)
	if err != nil {
		return "", errors.Wrap(err, "run command session")
	}

	sess = sess.Wait()

	for e := range sess.errors {
		err = errors.And(err, e)
	}

	return string(bytes.Trim(sess.Buffer().Contents(), "\n")), err
}

// Run - запускает команду оболочки с ожиданием выполнения
func Run(name string, command Command, writers ...io.Writer) error {
	sess, err := StartCmd(NewCommand(name, command), writers...)
	if err != nil {
		return errors.Wrap(err, "run command session")
	}

	sess.Wait()

	return nil
}

// StartCmd - запускает команду оболочки завернутую в управляющую обертку
func StartCmd(cmd *exec.Cmd, writers ...io.Writer) (*Session, error) {
	exited := make(chan struct{})

	session := &Session{
		cmd:      cmd,
		bout:     iorw.NewBuffer(),
		berr:     iorw.NewBuffer(),
		errors:   make(chan error, 1024),
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
