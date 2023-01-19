// nolint: dupl
package execs

import (
	"os"
	"time"

	"git.eth4.dev/golibs/errors"
)

type orderedGroup struct {
	members Members
	pool    map[string]Process
}

// NewOrdered - конструктор группы последовательного запуска раннеров
func NewOrdered(members ...Member) Runner {
	return &orderedGroup{
		members: members,
		pool:    make(map[string]Process),
	}
}

// Run - колбэк запуска раннера
func (g *orderedGroup) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	exitTrace := make(ExitTrace, 0, len(g.members))

	for m := 0; m < len(g.members); m++ {
		member := g.members[m]
		p := Background(member)
		g.pool[member.Name] = p

	Next:
		for {
			select {
			case <-p.Ready():
				break Next
			case err := <-p.Wait():
				exit := ExitEvent{Member: member}

				if err != nil {
					exit.Err = errors.Ctx().
						Str("runner", member.Name).
						Wrap(err, "exit runner status")
				}

				exitTrace = append(exitTrace, exit)

				break Next
			}
		}
	}

	close(ready)

	return g.wait(signals, exitTrace)
}

func (g *orderedGroup) wait(signals <-chan os.Signal, exitTrace ExitTrace) error {
	var exitErr error

	exited := map[string]struct{}{}
	signal := <-signals

	if len(exitTrace) > 0 {
		for _, exitEvent := range exitTrace {
			exited[exitEvent.Member.Name] = struct{}{}

			if exitEvent.Err != nil {
				errors.And(exitErr, exitEvent.Err)
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
