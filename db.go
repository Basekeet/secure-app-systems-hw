package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var input = bufio.NewScanner(os.Stdin)

func ask(label string) string {
	fmt.Print(label + ": ")
	if !input.Scan() {
		return ""
	}
	return input.Text()
}

func connectionConfig(user, password string) (*pgx.ConnConfig, error) {
	var file struct {
		Host     string
		Port     uint16
		Database string
	}
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	cfg, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}
	cfg.Host, cfg.Port, cfg.Database = file.Host, file.Port, file.Database
	cfg.User, cfg.Password = user, password
	cfg.ConnectTimeout = 5 * time.Second
	return cfg, nil
}

func connect(ctx context.Context, user, password string) (*pgx.Conn, error) {
	cfg, err := connectionConfig(user, password)
	if err != nil {
		return nil, err
	}
	return pgx.ConnectConfig(ctx, cfg)
}

func number(text string) (int, error) { return strconv.Atoi(strings.TrimSpace(text)) }
