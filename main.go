package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	user := ask("Username")
	password := ask("Password")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := connect(ctx, user, password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to the database")
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	var version string
	if err := conn.QueryRow(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		fmt.Fprintln(os.Stderr, "Could not get the database version")
		os.Exit(1)
	}
	fmt.Println("Connected successfully")
	fmt.Println(version)
}
