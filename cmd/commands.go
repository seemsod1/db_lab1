package main

import (
	"github.com/seemsod1/db_lab1/internal/handlers"
	"github.com/spf13/cobra"
)

func initRootCmd() *cobra.Command {

	var rootCmd = &cobra.Command{Use: "app"}
	var insertMCmd = &cobra.Command{
		Use:   "insert-m <id> <name> <mail> <age>",
		Short: "Insert master record into file",
		Long:  ``,
		Args:  cobra.ExactArgs(4),
		Run:   handlers.Repo.InsertM,
	}

	var getMCmd = &cobra.Command{
		Use:   "get-m <id>",
		Short: "Get master record from file",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		Run:   handlers.Repo.GetM,
	}
	var insertSCmd = &cobra.Command{
		Use:   "insert-s <user_id> <rent_id> <price> <country>",
		Short: "Insert slave record into file",
		Long:  ``,
		Args:  cobra.ExactArgs(4),
		Run:   handlers.Repo.InsertS,
	}
	var utilMCmd = &cobra.Command{
		Use:   "ut-m",
		Short: "Utility for master file",
		Long:  ``,
		Run:   handlers.Repo.UtilM,
	}
	var utilSCmd = &cobra.Command{
		Use:   "ut-s",
		Short: "Utility for slave file",
		Long:  ``,
		Run:   handlers.Repo.UtilS,
	}
	var getSCmd = &cobra.Command{
		Use:   "get-s <user_id> <rent_id>",
		Short: "Get slave record from file",
		Long:  ``,
		Args:  cobra.ExactArgs(2),
		Run:   handlers.Repo.GetS,
	}
	var updateMCmd = &cobra.Command{
		Use:   "update-m <id> <name> <mail> <age>",
		Short: "Update master record in file",
		Long:  ``,
		Args:  cobra.ExactArgs(4),
		Run:   handlers.Repo.UpdateM,
	}

	var updateSCmd = &cobra.Command{
		Use:   "update-s <user_id> <rent_id> <price> <country>",
		Short: "Update slave record in file",
		Long:  ``,
		Args:  cobra.ExactArgs(4),
		Run:   handlers.Repo.UpdateS,
	}

	var deleteSCmd = &cobra.Command{
		Use:   "delete-s <user_id> <rent_id>",
		Short: "Delete slave record from file",
		Long:  ``,
		Args:  cobra.ExactArgs(2),
		Run:   handlers.Repo.DeleteS,
	}

	rootCmd.AddCommand(insertMCmd)
	rootCmd.AddCommand(getMCmd)
	rootCmd.AddCommand(updateMCmd)
	rootCmd.AddCommand(utilMCmd)

	rootCmd.AddCommand(deleteSCmd)
	rootCmd.AddCommand(utilSCmd)
	rootCmd.AddCommand(getSCmd)
	rootCmd.AddCommand(insertSCmd)
	rootCmd.AddCommand(updateSCmd)

	return rootCmd

}
