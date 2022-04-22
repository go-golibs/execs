package execs

import "os"

// Process интерфейс процесса
type Process interface {
	// Ready синхронизация готовности процесса
	Ready() <-chan struct{}
	// Wait синхронизация завершения процесса
	Wait() <-chan error
	// Signal отправка системного сигнала процессу
	Signal(os.Signal)
}

type process struct {
	runner     Runner
	signals    chan os.Signal
	ready      chan struct{}
	exited     chan struct{}
	exitStatus error
}

// Background - конструтор процесса запущенного в бэкграутнд моде
func Background(r Runner) Process {
	p := newProcess(r)
	go p.run()

	return p
}

// Start - конструктор процесса
func Start(r Runner) Process {
	p := Background(r)

	select {
	case <-p.Ready():
	case <-p.Wait():
	}

	return p
}

func newProcess(runner Runner) *process {
	return &process{
		runner:  runner,
		signals: make(chan os.Signal),
		ready:   make(chan struct{}),
		exited:  make(chan struct{}),
	}
}

// Ready синхронизация готовности процесса
func (p *process) Ready() <-chan struct{} {
	return p.ready
}

// Wait синхронизация завершения процесса
func (p *process) Wait() <-chan error {
	exitChan := make(chan error, 1)

	go func() {
		<-p.exited
		exitChan <- p.exitStatus
	}()

	return exitChan
}

// Signal отправка системного сигнала процессу
func (p *process) Signal(signal os.Signal) {
	go func() {
		select {
		case p.signals <- signal:
		case <-p.exited:
		}
	}()
}

func (p *process) run() {
	p.exitStatus = p.runner.Run(p.signals, p.ready)
	close(p.exited)
}
