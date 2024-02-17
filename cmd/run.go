package main

import (
	"bufio"
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/seemsod1/db_lab1/internal/driver"
	"github.com/seemsod1/db_lab1/internal/driver/utils"
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
			//save master indexes
			utils.WriteIndices(driver.MasterFilename, app.Master.Ind)
			//save slave indexes
			utils.WriteIndices(driver.SlaveFilename, app.Slave.Ind)
			//close master file
			app.Master.FL.Close()
			//close slave file
			app.Slave.FL.Close()
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
