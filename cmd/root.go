package cmd

import (
	"fmt"
	"pantui/internal/tui"

	"github.com/spf13/cobra"
)

var (
	manifestURL string
	filePath    string
	versionInfo struct {
		version string
		commit  string
		date    string
	}
)

var rootCmd = &cobra.Command{
	Use:   "pantui [URL_OR_FILE]",
	Short: "A TUI for exploring HLS manifests",
	Long: `pantui is a terminal user interface for exploring HLS (HTTP Live Streaming) manifests.
Navigate through master manifests, media playlists, and segments with detailed FFProbe analysis.

Examples:
  pantui https://example.com/master.m3u8
  pantui /path/to/manifest.m3u8
  pantui --url https://example.com/master.m3u8
  pantui --file /path/to/manifest.m3u8`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetURL, targetFile string
		
		// Handle direct argument
		if len(args) == 1 {
			arg := args[0]
			if len(arg) >= 4 && arg[:4] == "http" {
				targetURL = arg
			} else {
				targetFile = arg
			}
		}
		
		// Override with flags if provided
		if manifestURL != "" {
			targetURL = manifestURL
			targetFile = ""
		}
		if filePath != "" {
			targetFile = filePath
			targetURL = ""
		}
		
		if targetURL == "" && targetFile == "" {
			return fmt.Errorf("please provide a manifest URL or file path")
		}
		
		app := tui.NewApp()
		
		if targetURL != "" {
			return app.RunWithURL(targetURL)
		} else {
			return app.RunWithFile(targetFile)
		}
	},
}

// SetVersionInfo sets the version information from main package
func SetVersionInfo(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&manifestURL, "url", "u", "", "HLS manifest URL")
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "Local HLS manifest file path")
	
	rootCmd.MarkFlagsMutuallyExclusive("url", "file")
}
