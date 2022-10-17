package api

import (
	"fmt"
	"os"

	"github.com/lmsilva-oss/albion-craft-ui/server/config"
	"github.com/lmsilva-oss/albion-craft-ui/server/src"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "api - starts the albion craft api",
	Long: `api is the backend for the albion craft UI
   
The CLI allows customization of initialization parameters for the API`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Load()
		src.Start()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
