package commands

import (
	"fmt"
	"os"

	"zbwrap/internal/registries"
	"zbwrap/internal/services"

	"github.com/spf13/cobra"
)

var syncDeep bool

var syncCmd = &cobra.Command{
	Use:   "sync [alias]",
	Short: "Synchronize repository metadata",
	Long:  `Scans the repository for missing metadata sidecars and regenerates them. Use --deep to analyze file contents.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		registry := registries.NewLocalRegistry()
		if err := registry.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			os.Exit(1)
		}

		repoPath, ok := registry.Get(alias)
		if !ok {
			fmt.Fprintf(os.Stderr, "Repository not found: %s\n", alias)
			os.Exit(1)
		}

		zbackupPath := registry.ZBackupPath
		if zbackupPath == "" {
			zbackupPath = "zbackup" // Default to PATH lookups if not configured
		}

		passwordFile := ""
		if registry.Encryption.Type == "password-file" {
			passwordFile = registry.Encryption.CredentialsPath
		}

		inspector := services.NewRepositoryInspector()
		if err := inspector.Sync(zbackupPath, repoPath, syncDeep, passwordFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing repository: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Synchronization complete for repository '%s'.\n", alias)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncDeep, "deep", false, "Perform deep inspection (MIME types)")
}
