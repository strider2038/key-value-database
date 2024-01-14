package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/strider2038/key-value-database/internal/config"
	"github.com/strider2038/key-value-database/internal/di"
)

func main() {
	options, err := config.LoadClientOptions()
	if err != nil {
		log.Fatalln("parse command line arguments: ", err)
	}

	client, err := di.NewClient(options)
	if err != nil {
		log.Fatalln("create client: ", err)
	}
	defer client.Close()

	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("command: ")
		command, _ := input.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "exit" || command == "\\q" {
			break
		}
		result, err := client.Send([]byte(command))
		if err != nil {
			fmt.Println("ERROR: ", err.Error())

			break
		}

		fmt.Println("result: ", string(result))
	}
}
