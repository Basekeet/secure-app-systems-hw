package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var tables = map[string][]string{
	"customers":   {"name", "email"},
	"products":    {"name", "price"},
	"orders":      {"customer_id", "status"},
	"order_items": {"order_id", "product_id", "quantity"},
}
var logFile *os.File

func report(message string, err error) {
	var stream io.Writer = os.Stdout
	if err != nil {
		stream = os.Stderr
		message = "Operation failed. Check your input and database availability."
	}
	fmt.Fprintln(stream, message)
	if logFile != nil {
		if err != nil {
			fmt.Fprintln(logFile, time.Now().Format(time.RFC3339), err)
		} else {
			fmt.Fprintln(logFile, time.Now().Format(time.RFC3339), message)
		}
	}
}

func column(table, name string, allowID bool) (string, error) {
	if _, ok := tables[table]; !ok {
		return "", fmt.Errorf("unknown table")
	}
	if allowID && name == "id" {
		return pgx.Identifier{name}.Sanitize(), nil
	}
	for _, allowed := range tables[table] {
		if name == allowed {
			return pgx.Identifier{name}.Sanitize(), nil
		}
	}
	return "", fmt.Errorf("unknown or forbidden column")
}

func tableName() (string, error) {
	name := ask("Table: customers / products / orders / order_items")
	if _, ok := tables[name]; !ok {
		return "", fmt.Errorf("unknown table")
	}
	return name, nil
}
func fields(table string, allowID bool, start int, separator string) (string, []any, error) {
	count, err := number(ask("Number of fields"))
	if err != nil || count < 1 || count > 10 {
		return "", nil, fmt.Errorf("invalid field count")
	}
	parts := []string{}
	values := []any{}
	seen := map[string]bool{}
	for i := 0; i < count; i++ {
		name := ask("Column")
		col, err := column(table, name, allowID)
		if err != nil || seen[name] {
			return "", nil, fmt.Errorf("invalid or duplicate column")
		}
		seen[name] = true
		parts = append(parts, fmt.Sprintf("%s = $%d", col, start+i))
		values = append(values, ask("Value"))
	}
	return strings.Join(parts, separator), values, nil
}

func view(ctx context.Context, conn *pgx.Conn, filtered bool) error {
	table, err := tableName()
	if err != nil {
		return err
	}
	query := "SELECT * FROM " + pgx.Identifier{table}.Sanitize()
	var args []any
	if filtered {
		condition, values, err := fields(table, true, 1, " AND ")
		if err != nil {
			return err
		}
		query += " WHERE " + condition
		args = values
	}
	query += " ORDER BY id"
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for _, field := range rows.FieldDescriptions() {
		fmt.Printf("%s\t", field.Name)
	}
	fmt.Println()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return err
		}
		for _, value := range values {
			if price, ok := value.(pgtype.Numeric); ok {
				value, err = price.Value()
				if err != nil {
					return err
				}
			}
			fmt.Printf("%v\t", value)
		}
		fmt.Println()
	}
	return rows.Err()
}

func update(ctx context.Context, conn *pgx.Conn, multiple bool) error {
	table, err := tableName()
	if err != nil {
		return err
	}
	set, args, err := fields(table, false, 1, ", ")
	if err != nil {
		return err
	}
	var condition string
	if !multiple {
		id, err := number(ask("Record ID"))
		if err != nil {
			return err
		}
		args = append(args, id)
		condition = fmt.Sprintf("id = $%d", len(args))
	} else {
		col, err := column(table, ask("Filter column"), true)
		if err != nil {
			return err
		}
		n, err := number(ask("Number of IN values"))
		if err != nil || n < 1 || n > 100 {
			return fmt.Errorf("invalid count")
		}
		placeholders := []string{}
		for i := 0; i < n; i++ {
			args = append(args, ask("Filter value"))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		condition = col + " IN (" + strings.Join(placeholders, ", ") + ")"
	}
	tag, err := conn.Exec(ctx, "UPDATE "+pgx.Identifier{table}.Sanitize()+" SET "+set+" WHERE "+condition, args...)
	if err == nil {
		fmt.Println("Rows updated:", tag.RowsAffected())
	}
	return err
}

func insert(ctx context.Context, conn *pgx.Conn, multiple bool) error {
	table, err := tableName()
	if err != nil {
		return err
	}
	count := 1
	if multiple {
		count, err = number(ask("Number of rows"))
		if err != nil || count < 1 || count > 100 {
			return fmt.Errorf("invalid count")
		}
	}
	names := []string{}
	for _, name := range tables[table] {
		names = append(names, pgx.Identifier{name}.Sanitize())
	}
	args := []any{}
	groups := []string{}
	for row := 0; row < count; row++ {
		fmt.Println("Row", row+1)
		placeholders := []string{}
		for _, name := range tables[table] {
			args = append(args, ask(name))
			placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
		}
		groups = append(groups, "("+strings.Join(placeholders, ", ")+")")
	}
	_, err = conn.Exec(ctx, "INSERT INTO "+pgx.Identifier{table}.Sanitize()+" ("+strings.Join(names, ", ")+") VALUES "+strings.Join(groups, ", "), args...)
	return err
}
func linked(ctx context.Context, conn *pgx.Conn, multiple bool) error {
	count := 1
	var err error
	if multiple {
		count, err = number(ask("Number of orders"))
		if err != nil || count < 1 || count > 100 {
			return fmt.Errorf("invalid count")
		}
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	for i := 0; i < count; i++ {
		fmt.Println("Order", i+1)
		customer, err := number(ask("Customer ID"))
		if err != nil {
			return err
		}
		status := ask("Status")
		items := 1
		if multiple {
			items, err = number(ask("Number of items"))
			if err != nil || items < 1 || items > 100 {
				return fmt.Errorf("invalid count")
			}
		}
		products := make([]int, items)
		quantities := make([]int, items)
		for j := 0; j < items; j++ {
			products[j], err = number(ask("Product ID"))
			if err != nil {
				return err
			}
			quantities[j], err = number(ask("Quantity"))
			if err != nil {
				return err
			}
		}
		var id int
		if err = tx.QueryRow(ctx, "INSERT INTO orders (customer_id,status) VALUES ($1,$2) RETURNING id", customer, status).Scan(&id); err != nil {
			return err
		}
		args := []any{}
		groups := []string{}
		for j := 0; j < items; j++ {
			base := len(args)
			args = append(args, id, products[j], quantities[j])
			groups = append(groups, fmt.Sprintf("($%d,$%d,$%d)", base+1, base+2, base+3))
		}
		if _, err = tx.Exec(ctx, "INSERT INTO order_items (order_id,product_id,quantity) VALUES "+strings.Join(groups, ","), args...); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func main() {
	if path := os.Getenv("LOG_FILE"); path != "" {
		var err error
		logFile, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			report("", err)
			return
		}
		defer logFile.Close()
	}
	user, password := ask("Username"), ask("Password")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := connect(ctx, user, password)
	cancel()
	if err != nil {
		report("", err)
		return
	}
	defer conn.Close(context.Background())
	report("Connected successfully", nil)
	if _, err := conn.Exec(context.Background(), "SET statement_timeout = '5s'"); err != nil {
		report("", err)
		return
	}
	for {
		fmt.Println("1 View all; 2 Filter; 3 Update one; 4 Update multiple; 5 Insert one; 6 Order with item; 7 Insert multiple; 8 Orders with items; 0 Exit")
		choice := ask("Command")
		if choice == "0" || choice == "" {
			return
		}
		ctx := context.Background()
		switch choice {
		case "1":
			err = view(ctx, conn, false)
		case "2":
			err = view(ctx, conn, true)
		case "3":
			err = update(ctx, conn, false)
		case "4":
			err = update(ctx, conn, true)
		case "5":
			err = insert(ctx, conn, false)
		case "6":
			err = linked(ctx, conn, false)
		case "7":
			err = insert(ctx, conn, true)
		case "8":
			err = linked(ctx, conn, true)
		default:
			fmt.Println("Choose a menu option")
			continue
		}
		report("Operation completed", err)
	}
}
