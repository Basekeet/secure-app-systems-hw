package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	interval, err := time.ParseDuration(os.Getenv("CHECK_INTERVAL"))
	if err != nil || interval <= 0 {
		fmt.Fprintln(os.Stderr, "CHECK_INTERVAL must be positive, for example 5m")
		os.Exit(1)
	}
	var out io.Writer = os.Stdout
	var errors io.Writer = os.Stderr
	if path := os.Getenv("LOG_FILE"); path != "" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not open the log file")
			os.Exit(1)
		}
		defer file.Close()
		out = io.MultiWriter(os.Stdout, file)
		errors = io.MultiWriter(os.Stderr, file)
	}
	check := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn, err := connect(ctx, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))
		if err != nil {
			fmt.Fprintln(errors, time.Now().Format(time.RFC3339), "Connection failed:", err)
			return
		}
		defer conn.Close(context.Background())
		var version string
		if err := conn.QueryRow(ctx, "SELECT VERSION()").Scan(&version); err != nil {
			fmt.Fprintln(errors, time.Now().Format(time.RFC3339), "Query failed:", err)
			return
		}
		fmt.Fprintln(out, time.Now().Format(time.RFC3339), "Connected successfully:", version)
	}
	check()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		check()
	}
}
