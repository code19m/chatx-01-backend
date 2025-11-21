package main

import (
	"chatx-01-backend/internal/app"
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "http", "createsuperuser", "consume":
		run(command)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func run(command string) {
	ctx := context.Background()

	application, err := app.Build(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()

	switch command {
	case "http":
		if err := application.RunHTTPServer(); err != nil {
			log.Fatal(err)
		}
	case "createsuperuser":
		if err := application.CreateSuperUser(); err != nil {
			log.Fatal(err)
		}
	case "consume":
		if err := application.RunNotificationConsumer(); err != nil {
			log.Fatal(err)
		}
	}
}

func printUsage() {
	fmt.Println("Usage: chatx <command>")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  http              Start HTTP server")
	fmt.Println("  createsuperuser   Create a super user (admin)")
	fmt.Println("  consume           Start notification consumer service")
}
