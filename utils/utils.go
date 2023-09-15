package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func Signal() chan os.Signal {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	return sigCh
}
