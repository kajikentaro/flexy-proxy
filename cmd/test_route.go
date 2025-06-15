package cmd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/kajikentaro/flexy-proxy/utils"
	"gopkg.in/yaml.v3"
)

func TestRoute(customConfigPath string, testUrl string) error {
	proxyConfig, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		return err
	}

	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		return err
	}

	route, err := proxyConfig.Router.GetMatchedRoute(parsedUrl)
	if errors.Is(err, models.ErrRouteNotFound) {
		return fmt.Errorf("route not found")
	} else if err != nil {
		return err
	}

	yaml, err := yaml.Marshal(route)
	if err != nil {
		return err
	}
	fmt.Printf("Matched Route YAML:\n\n%s\n", string(yaml))

	return nil
}
