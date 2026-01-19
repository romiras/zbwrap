package commands

import (
	"fmt"
	"os"

	"zbwrap/internal/registries"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [alias] [path]",
	Short: "Add a new ZBackup repository",
	Long:  `Registers a new ZBackup repository with the given alias and filesystem path.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		path := args[1]

		registry := registries.NewLocalRegistry()
		if err := registry.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			os.Exit(1)
		}

		if err := registry.Add(alias, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding repository: %v\n", err)
			os.Exit(1)
		}

		if err := registry.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving registry: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Repository '%s' added successfully pointing to %s\n", alias, path)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
