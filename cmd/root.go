package cmd

import (
	"fmt"
	"pantui/internal/tui"

	"github.com/spf13/cobra"
)

var (
	manifestURL string
	filePath    string
)

var rootCmd = &cobra.Command{
	Use:   "pantui",
	Short: "A TUI for exploring HLS manifests",
	Long: `pantui is a terminal user interface for exploring HLS (HTTP Live Streaming) manifests.
Similar to k9s for Kubernetes, pantui provides an interactive way to navigate through
HLS manifests, sub-manifests, and chunks with colorized output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if manifestURL == "" && filePath == "" {
			return fmt.Errorf("either --url or --file must be specified")
		}
		
		app := tui.NewApp()
		
		if manifestURL != "" {
			return app.RunWithURL(manifestURL)
		} else {
			return app.RunWithFile(filePath)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&manifestURL, "url", "u", "", "HLS manifest URL")
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Local HLS manifest file path")
	
	rootCmd.MarkFlagsMutuallyExclusive("url", "file")
}
