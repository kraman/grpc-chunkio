package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	context "golang.org/x/net/context"
)

var rootCmd = &cobra.Command{
	Use:   "example",
	Short: "gRPC file streaming example",
}

var fetchCmd = &cobra.Command{
	Use:   "get",
	Short: "fetch file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return downloadFile(args[0], args[1])
	},
}

var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listen(context.Background(), ":9090")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(serveCmd)
}
