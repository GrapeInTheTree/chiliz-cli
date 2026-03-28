package cmd

import (
	"fmt"
	"os"

	"encoding/json"
	"github.com/spf13/cobra"
)

// Set via ldflags at build time by GoReleaser
var (
	version = "dev"
	commit  = "none"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print chiliz version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(map[string]string{
				"version": version,
				"commit":  commit,
			})
		} else {
			fmt.Printf("chiliz %s (%s)\n", version, commit)
		}
	},
}
