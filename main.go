package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/kajikentaro/flexy-proxy/models"
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

func startProxy(customConfigPath string, portNum int) {
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

func testRoute(customConfigPath string, testUrl string) error {
	proxyConfig, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		return err
	}

	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return err
	}

	handler, matchedUrl, err := proxyConfig.Router.GetRoundTripper(parsedUrl)

	if errors.Is(err, models.ErrRouteNotFound) {
		return fmt.Errorf("route not found")
	}
	if err != nil {
		return err
	}

	fmt.Println("URL:", matchedUrl)
	fmt.Println("Type:", handler.GetType())
	fmt.Println("Info:")
	for key, val := range handler.GetResponseInfo() {
		fmt.Printf("    %s: %s\n", key, val)
	}
	return nil
}

func main() {
	var (
		customConfigPath string
		portNum          int
	)

	rootCmd := &cobra.Command{
		Use:   "main",
		Short: "A YAML-based flexible proxy for software development",
		Run: func(cmd *cobra.Command, args []string) {
			startProxy(customConfigPath, portNum)
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
			startProxy(customConfigPath, portNum)
		},
	}

	testCmd := &cobra.Command{
		Use:   "test [URL]",
		Short: "Print the configured routing information for the given URL",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := testRoute(customConfigPath, args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(versionCmd, startCmd, testCmd)
	rootCmd.Execute()
}
