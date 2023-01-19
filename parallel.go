// nolint: dupl
package execs

import (
	"os"
	"sync"
	"time"

	"git.eth4.dev/golibs/errors"
)

type parallelGroup struct {
	members Members
	pool    map[string]Process
}

// NewParallel конструктор параллельного запуска раннеров группы
func NewParallel(members ...Member) Runner {
	return &parallelGroup{
		members: members,
		pool:    make(map[string]Process),
	}
}

// Run коллбэк запуска раннера группы
func (g parallelGroup) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	exitTrace := make(ExitTrace, 0, len(g.members))

	var (
		wg     sync.WaitGroup
		traces = make(chan ExitEvent, len(g.members))
	)

	for m := 0; m < len(g.members); m++ {
		wg.Add(1)

		member := g.members[m]
		p := Background(member)
		g.pool[member.Name] = p

		go func() {
		Done:
			for {
				select {
				case <-p.Ready():
					break Done
				case err := <-p.Wait():
					exit := ExitEvent{Member: member}

					if err != nil {
						exit.Err = errors.Ctx().
							Str("runner", member.Name).
							Wrap(err, "exit runner status")
					}

					traces <- exit

					break Done
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(traces)
	close(ready)

	for event := range traces {
		exitTrace = append(exitTrace, event)
	}

	return g.wait(signals, exitTrace)
}

func (g *parallelGroup) wait(signals <-chan os.Signal, exitTrace ExitTrace) error {
	var exitErr error

	exited := map[string]struct{}{}
	signal := <-signals

	if len(exitTrace) > 0 {
		for _, exitEvent := range exitTrace {
			exited[exitEvent.Member.Name] = struct{}{}

			if exitEvent.Err != nil {
				exitErr = errors.And(exitErr, exitEvent.Err)
			}
		}
	}

	for m := len(g.members) - 1; m >= 0; m-- {
		member := g.members[m]
		if _, isExited := exited[member.Name]; isExited {
			continue
		}

		if p, ok := g.pool[member.Name]; ok {
			p.Signal(signal)
		Exited:
			for {
				select {
				case err := <-p.Wait():
					if err != nil {
						exitErr = errors.And(
							exitErr,
							errors.Ctx().
								Str("runner", member.Name).
								Wrap(err, "exit runner status"),
						)
					}
					break Exited
				case <-time.After(time.Millisecond * 100):
				}
			}
		}
	}

	return exitErr
}
