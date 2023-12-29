package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/strider2038/key-value-database/internal/di"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	db := di.NewDatabase(di.Options{Logger: logger})

	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("command: ")
		command, _ := input.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "exit" || command == "\\q" {
			break
		}
		result, err := db.Execute(context.Background(), command)
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
		} else {
			fmt.Println("result: ", result)
		}
	}
}
