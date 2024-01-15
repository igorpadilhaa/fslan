package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/igorpadilhaa/fslan/client"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("syntax:", args[0], "<server name>")
		return
	}

	client, err := client.NewClientHTTP(args[1], 5173)
	if err != nil {
		fmt.Println("error: failed to launch client: ", err)
	}
	commands(client)
}

func commands(client client.Client) {
	shouldExit := false
	for !shouldExit {
		fmt.Print("> ")
		commands, err := readCommands()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		shouldExit, err = execCommand(commands, client)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func readCommands() ([]string, error) {
	reader := bufio.NewReader(os.Stdin)

	bytes, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	line := string(bytes)
	var commands []string

	for _, slice := range strings.Split(line, " ") {
		if len(slice) == 0 {
			continue
		}
		commands = append(commands, slice)
	}

	return commands, nil
}

func execCommand(command []string, client client.Client) (bool, error) {
	shouldExit := false
	if len(command) == 0 {
		return false, nil
	}

	switch command[0] {
	case "list":
		if len(command) == 1 {
			files := client.Shared()
			if len(files) == 0 {
				fmt.Println("no files is being shared")
				break
			}

			fmt.Println("Shared files:")
			for _, file := range files {
				fmt.Println(file.Name, file.Path)
			}
			break
		} else if len(command) > 2 {
			return false, errors.New("command syntax: list [host name]")
		}

		host := command[1]
		files, err := client.ListFiles(host)
		if err != nil {
			return false, err
		}

		fmt.Printf("%s shared files:\n", host)
		for _, file := range files {
			fmt.Println(file.Name, file.Mime)
		}

	case "share":
		if len(command) != 3 {
			return false, errors.New("command syntax: share <file path> <public name>")
		}

		path, name := command[1], command[2]
		client.ShareFile(name, path)

	case "get":
		if len(command) != 3 {
			return false, errors.New("command syntax: get <host> <file name>")
		}

		host := command[1]
		filename := command[2]

		err := client.GetFile(host, filename)
		if err != nil {
			return false, fmt.Errorf("failed to download file: %w", err)
		}

	case "servers":
		servers := client.Hosts()
		if len(servers) == 0 {
			fmt.Println("no server found")
			break
		}

		for _, server := range servers {
			fmt.Println(server)
		}

	case "exit":
		shouldExit = true

	default:
		return false, fmt.Errorf("unknown command %q", command[0])
	}
	return shouldExit, nil
}
