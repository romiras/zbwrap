package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"zbwrap/internal/registries"
	"zbwrap/internal/services"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var infoJson bool

var infoCmd = &cobra.Command{
	Use:   "info [alias]",
	Short: "Show repository details",
	Long:  `Displays detailed information about a specific repository, including disk usage and backup history.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		registry := registries.NewLocalRegistry()
		if err := registry.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
			os.Exit(1)
		}

		path, ok := registry.Get(alias)
		if !ok {
			fmt.Fprintf(os.Stderr, "Repository not found: %s\n", alias)
			os.Exit(1)
		}

		inspector := services.NewRepositoryInspector()
		details, err := inspector.Inspect(alias, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error inspecting repository: %v\n", err)
			os.Exit(1)
		}

		if infoJson {
			output, err := json.MarshalIndent(details, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		} else {
			printHumanReadable(details)
		}
	},
}

func printHumanReadable(details *services.RepoDetails) {
	fmt.Printf("REPOSITORY: %s [%s]\n", details.Alias, details.PhysicalPath)
	fmt.Printf("TOTAL DISK USAGE: %s\n", humanize.Bytes(uint64(details.TotalSizeBytes)))
	fmt.Println("-------------------------------------------------------------------------------------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "BACKUP NAME\tDATE\tMIME TYPE\tDESCRIPTION")
	fmt.Fprintln(w, "-------------------------------------------------------------------------------------------------")

	for _, b := range details.Backups {
		dateStr := b.Date.Format("Jan 02, 15:04")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.Filename, dateStr, b.MimeType, b.Description)
	}
	w.Flush()
	fmt.Println("")
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolVarP(&infoJson, "json", "j", false, "Output in JSON format")
}
