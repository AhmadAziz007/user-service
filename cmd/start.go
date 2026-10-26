package cmd

import (
	"ecommerce/user-service/internal/app"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start",
	Long:  "start",
	Run: func(cmd *cobra.Command, args []string) {
		// call func route api
		app.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
