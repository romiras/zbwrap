package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup [alias]",
	Short: "Create a new backup",
	Long:  `Create a new backup for the repository associated with the given alias.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoAlias := args[0]
		fmt.Printf("Starting backup for alias: %s\n", repoAlias)
		// Logic to trigger backup service would go here
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
