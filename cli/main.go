/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"

	"cli/cmd"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

	cmd.Execute()
}
