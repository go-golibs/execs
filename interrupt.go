package execs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// InterruptHandler - запускает обработчик сигналов прерывания программы
// и возвращает контекст с синхронизацией завершения работы
func InterruptHandler(ctx context.Context, eh ...func()) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(ctx)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig

		_, _ = fmt.Fprintln(os.Stdout, "\r- Ctrl+C pressed in Terminal")

		if len(eh) == 0 {
			cancel()

			return
		}

		for _, apply := range eh {
			apply()
		}
	}()

	return ctx
}
