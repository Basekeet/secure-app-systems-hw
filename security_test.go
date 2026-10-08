package main

import "testing"

func TestColumnRestrictions(t *testing.T) {
	for _, name := range []string{"id", "name; DROP TABLE customers", "id OR 1=1"} {
		if _, err := column("customers", name, false); err == nil {
			t.Fatalf("accepted forbidden update column %q", name)
		}
	}
	if _, err := column("customers", "id", true); err != nil {
		t.Fatal(err)
	}
	if _, err := column("unknown", "id", true); err == nil {
		t.Fatal("accepted unknown table")
	}
}

func TestCredentialsAreLiteral(t *testing.T) {
	user := "student sslmode=require host=evil"
	password := "pass dbname=other"
	cfg, err := connectionConfig(user, password)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User != user || cfg.Password != password || cfg.Host != "db" || cfg.Database != "homework" {
		t.Fatal("credentials changed connection options")
	}
}
