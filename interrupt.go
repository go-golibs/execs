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
func InterruptHandler(ctx context.Context, handles ...func()) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(ctx)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Println("\r- Ctrl+C pressed in Terminal")

		for _, handle := range handles {
			handle()
		}

		cancel()
	}()

	return ctx
}
