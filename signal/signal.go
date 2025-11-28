package signal

import (
	"os"
	"os/signal"
	"syscall"
)

type handler func()

// Stop handles term and int signals
func Stop(h handler) chan int {
	closed := make(chan int)

	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint
		h()
		close(closed)
	}()
	return closed
}
