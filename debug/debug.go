package debug

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"

	"github.com/Sirupsen/logrus"
)

func init() {
	debugC := make(chan os.Signal, 1)
	signal.Notify(debugC, syscall.SIGUSR1)
	ftrace := filepath.Join(os.TempDir(), "swan-stack-trace.log")

	go func() {
		for range debugC {
			f, err := os.OpenFile(ftrace, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				logrus.Error("write stack trace log file error", err)
				break
			}
			fmt.Fprint(f, "GOROUTINE\n\n")
			pprof.Lookup("goroutine").WriteTo(f, 2)
			fmt.Fprint(f, "\n\nHEAP\n\n")
			pprof.Lookup("heap").WriteTo(f, 1)
			fmt.Fprint(f, "\n\nTHREADCREATE\n\n")
			pprof.Lookup("threadcreate").WriteTo(f, 1)
			fmt.Fprint(f, "\n\nBLOCK\n\n")
			pprof.Lookup("block").WriteTo(f, 1)
			f.Close()
		}
	}()
}
