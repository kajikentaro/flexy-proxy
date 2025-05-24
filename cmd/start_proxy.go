package cmd

import (
	"fmt"
	"net/http"

	"github.com/kajikentaro/flexy-proxy/proxy"
	"github.com/kajikentaro/flexy-proxy/utils"
)

func StartProxy(customConfigPath string, portNum int) error {
	proxyConfig, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		return err
	}

	proxy := proxy.SetupProxy(proxyConfig)
	addr := fmt.Sprintf(":%d", portNum)
	proxyConfig.Logger.Info(fmt.Sprintf("Proxy started on %s", addr))
	return http.ListenAndServe(addr, proxy)
}
