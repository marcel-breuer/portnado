package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/marcel-breuer/portnado/internal/cli"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	app, err := cli.New(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(app.Run(ctx, os.Args[1:]))
}
