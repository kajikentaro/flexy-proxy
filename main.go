package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kajikentaro/flexy-proxy/proxy"
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

var (
	customConfigPath string
	portNum          int
)

func startProxy() {
	proxyConfig, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	proxy := proxy.SetupProxy(proxyConfig)
	addr := fmt.Sprintf(":%d", portNum)
	proxyConfig.Logger.Info(fmt.Sprintf("Proxy started on %s", addr))
	fmt.Fprintf(os.Stderr, "%v\n", http.ListenAndServe(addr, proxy))
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A YAML-based flexible proxy for software development",
		Run: func(cmd *cobra.Command, args []string) {
			startProxy()
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
		Run: func(cmd *cobra.Command, args []string) {
			startProxy()
		},
	}

	rootCmd.AddCommand(versionCmd, startCmd)
	rootCmd.Execute()
}
