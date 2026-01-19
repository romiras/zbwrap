package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"zbwrap/internal/registries"

	"github.com/spf13/cobra"
)

var listJson bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List managed repositories",
	Long:  `Lists all ZBackup repositories currently managed by zbwrap.`,
	Run: func(cmd *cobra.Command, args []string) {
		registry := registries.NewLocalRegistry()
		if err := registry.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			os.Exit(1)
		}

		repos := registry.List()

		if listJson {
			output, err := json.MarshalIndent(repos, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		} else {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ALIAS\tPATH")
			for alias, path := range repos {
				fmt.Fprintf(w, "%s\t%s\n", alias, path)
			}
			w.Flush()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listJson, "json", false, "Output in JSON format")
}
