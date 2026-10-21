// Package baseconfig holds the configuration every service in this repository
// needs, and a loader that fills a struct from environment variables.
//
// Services embed Base in their own config struct and add fields with
// envconfig tags. See internal/config for the pattern.
package baseconfig

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Base is the set of environment variables common to every service.
type Base struct {
	// AppName identifies the service in logs and telemetry resource attributes.
	AppName string `envconfig:"APP_NAME" required:"true"`
	// Host is the interface to bind. Empty binds every interface.
	Host string `envconfig:"HOST" default:""`
	// Port is the TCP port to listen on.
	Port int `envconfig:"PORT" default:"8000"`
	// Environment names the deployment target: local, test, development,
	// staging or production.
	Environment string `envconfig:"ENVIRONMENT" default:"local"`
	// LogLevel is one of debug, info, warn or error.
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load fills cfg from the process environment. Fields are matched by their
// envconfig tag, defaults apply when a variable is unset, and a missing
// required variable is an error.
func Load[T any](cfg *T) error {
	if err := envconfig.Process("", cfg); err != nil {
		return fmt.Errorf("loading configuration from environment: %w", err)
	}
	return nil
}
