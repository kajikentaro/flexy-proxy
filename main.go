package main

import (
	"fmt"
	"os"

	"github.com/kajikentaro/flexy-proxy/cmd"
	"github.com/kajikentaro/flexy-proxy/utils"
	"github.com/spf13/cobra"
)

// this will be specified like:
// go build -ldflags "-X 'main.version=1.0.0'"
var version string

func getVersion() string {
	if version == "" {
		return "unknown"
	}
	return version
}

func run(runner func() error) {
	if err := runner(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	var (
		customConfigPath string
		portNum          int
	)

	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A YAML-based flexible proxy for software development",
		Run: func(c *cobra.Command, args []string) {
			run(func() error {
				return cmd.StartProxy(customConfigPath, portNum)
			})
		},
	}

	rootCmd.PersistentFlags().StringVarP(&customConfigPath, "config", "f", utils.DEFAULT_CONFIG_PATH, "Path to custom config file")
	rootCmd.PersistentFlags().IntVarP(&portNum, "port", "p", 8888, "Port number")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(getVersion())
		},
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy server",
		Run: func(c *cobra.Command, args []string) {
			run(func() error {
				return cmd.StartProxy(customConfigPath, portNum)
			})
		},
	}

	testCmd := &cobra.Command{
		Use:   "test [URL]",
		Short: "Print the configured routing information for the given URL",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			run(func() error {
				return cmd.TestRoute(customConfigPath, args[0])
			})
		},
	}

	rootCmd.AddCommand(versionCmd, startCmd, testCmd)
	rootCmd.Execute()
}
