// Command playground is a local validation UI. It is not required by the SDK.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/xinhaolin/htmlToWord/tools/playground/internal/server"
)

func main() {
	address := flag.String("addr", "127.0.0.1:8787", "loopback listen address")
	batch := flag.String("batch-dir", "", "optional validation artifact directory")
	flag.Parse()
	if *batch != "" {
		os.Setenv("HTMLWORD_BATCH_DIR", *batch)
	} else if os.Getenv("HTMLWORD_BATCH_DIR") == "" {
		absolute, _ := filepath.Abs("../validation/artifacts/batch")
		os.Setenv("HTMLWORD_BATCH_DIR", absolute)
	}
	logger := log.New(os.Stderr, "htmlword-playground ", log.LstdFlags)
	app := server.New(logger)
	defer app.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	service := &http.Server{Addr: *address, Handler: app.Handler(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 100 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		service.Shutdown(shutdown)
	}()
	logger.Printf("Validation playground: http://%s", *address)
	if err := service.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Print(err)
	}
}
