package commands

import (
	"fmt"
	"os"
	"zbwrap/internal/registries"
	"zbwrap/internal/services"

	"github.com/spf13/cobra"
)

var (
	backupSuffix      string
	backupDescription string
)

var backupCmd = &cobra.Command{
	Use:   "backup [alias]",
	Short: "Create a new backup",
	Long:  `Create a new backup for the repository associated with the given alias.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoAlias := args[0]

		registry := registries.NewLocalRegistry()
		if err := registry.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			os.Exit(1)
		}

		repoPath, ok := registry.Get(repoAlias)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: repository alias '%s' not found\n", repoAlias)
			os.Exit(1)
		}

		runner := services.NewBackupRunner(registry)

		fmt.Printf("Starting backup for alias: %s (%s)\n", repoAlias, repoPath)

		// Stream from stdin to zbackup
		if err := runner.Backup(repoPath, backupSuffix, backupDescription, os.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Backup completed successfully.")
	},
}

func init() {
	backupCmd.Flags().StringVarP(&backupSuffix, "suffix", "s", "manual", "suffix for the backup filename")
	backupCmd.Flags().StringVarP(&backupDescription, "description", "m", "", "optional description for the backup")
	rootCmd.AddCommand(backupCmd)
}
