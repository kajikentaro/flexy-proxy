package version

import (
	"fmt"
	"strconv"
	"strings"
)

// this will be specified like:
// go build -ldflags "-X $(go list -f '{{.ImportPath}}' ./utils/version).version=v1.0.0"
var version string

func GetVersion() string {
	if version == "" {
		return "unknown"
	}
	return version
}

func VerifyVersion(required string, current string) error {
	if required == "" || current == "" || current == "unknown" {
		return nil
	}

	parseVersion := func(version, name string) ([]int, error) {
		splitted := strings.Split(version[1:], ".")

		if len(splitted) != 3 || version[:1] != "v" {
			return nil, fmt.Errorf("format of %s must be \"v[major].[minor].[patch]\"", name)
		}

		res := make([]int, 3)
		for i, s := range splitted {
			num, err := strconv.Atoi(s)
			if err != nil {
				return nil, fmt.Errorf("invalid %s \"%s\". failed to parse \"%s\" to int", name, version, s)
			}
			res[i] = num
		}

		return res, nil
	}

	rNumbers, err := parseVersion(required, "required version")
	if err != nil {
		return err
	}
	cNumbers, err := parseVersion(current, "build version")
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		rr := rNumbers[i]
		cc := cNumbers[i]

		if cc == rr {
			continue
		}

		if cc < rr {
			return fmt.Errorf("Flexy Proxy doesn't support this configuration. Please upgrade Flexy Proxy to %s or later", required)
		}

		if cc > rr {
			return nil
		}
	}

	return nil
}
