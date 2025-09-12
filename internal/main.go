package main

import (
	"context"
	"fmt"
	"internal/db"
	"os"
)

func main() {
	conn := db.ConnectOrFail()
	defer conn.Close()

	var greeting string
	err := conn.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

}
