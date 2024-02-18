package main

import (
	"bufio"
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func run(rootCmd *cobra.Command, reader *bufio.Reader) error {
	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading command: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("Exiting...")
			return nil
		}

		args, err := shellquote.Split(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing command: %v\n", err)
			continue
		}

		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
		}
	}
}
