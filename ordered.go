// nolint: dupl
package execs

import (
	"os"
	"time"

	"git.corout.in/golibs/errors"
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
	errTrace := make(ErrorTrace, 0, len(g.members))

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
				errTrace = append(errTrace, ExitEvent{
					Member: member,
					Err:    errors.Ctx().Str("member", member.Name).Just(err),
				})
				break Next
			}
		}
	}

	close(ready)

	return g.wait(signals, errTrace)
}

func (g *orderedGroup) wait(signals <-chan os.Signal, errTrace ErrorTrace) error {
	errOccurred := false
	exited := map[string]struct{}{}
	signal := <-signals

	if len(errTrace) > 0 {
		for _, exitEvent := range errTrace {
			exited[exitEvent.Member.Name] = struct{}{}

			if exitEvent.Err != nil {
				errOccurred = true
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
						errTrace = append(errTrace, ExitEvent{
							Member: member,
							Err:    errors.Ctx().Str("member", member.Name).Just(err),
						})
						errOccurred = true
					}
					break Exited
				case <-time.After(time.Millisecond * 100):
				}
			}
		}
	}

	if errOccurred {
		return errTrace
	}

	return nil
}
