package main

import (
	"context"
	"fmt"
	"github.com/aleksey-kombainov/gophermart-sp.git/internal/app"
	"os"
)

func main() {
	//ctx, cancel := signal.NotifyContext(
	//	context.Background(),
	//	syscall.SIGINT,
	//	syscall.SIGTERM,
	//)
	//defer cancel()
	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		app.Shutdown(app.ExitCodeErrorGeneral)
	}
}
