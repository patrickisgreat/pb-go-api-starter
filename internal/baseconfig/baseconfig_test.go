package baseconfig

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
)

type testConfig struct {
	Base

	Extra string `envconfig:"EXTRA" default:"fallback"`
}

// unsetenv removes name for the duration of the test and restores it afterwards.
func unsetenv(t *testing.T, name string) {
	t.Helper()
	t.Setenv(name, "") // registers cleanup that restores the original value
	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	c := qt.New(t)
	t.Setenv("APP_NAME", "svc")
	for _, name := range []string{"HOST", "PORT", "ENVIRONMENT", "LOG_LEVEL", "EXTRA"} {
		unsetenv(t, name)
	}

	var cfg testConfig
	c.Assert(Load(&cfg), qt.IsNil)
	c.Assert(cfg.AppName, qt.Equals, "svc")
	c.Assert(cfg.Port, qt.Equals, 8000)
	c.Assert(cfg.Environment, qt.Equals, "local")
	c.Assert(cfg.LogLevel, qt.Equals, "info")
	c.Assert(cfg.Extra, qt.Equals, "fallback")
}

func TestLoadReadsValues(t *testing.T) {
	c := qt.New(t)
	t.Setenv("APP_NAME", "svc")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "9000")
	t.Setenv("ENVIRONMENT", "test")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("EXTRA", "set")

	var cfg testConfig
	c.Assert(Load(&cfg), qt.IsNil)
	c.Assert(cfg.Host, qt.Equals, "127.0.0.1")
	c.Assert(cfg.Port, qt.Equals, 9000)
	c.Assert(cfg.Environment, qt.Equals, "test")
	c.Assert(cfg.LogLevel, qt.Equals, "debug")
	c.Assert(cfg.Extra, qt.Equals, "set")
}

func TestLoadRequiresAppName(t *testing.T) {
	c := qt.New(t)
	unsetenv(t, "APP_NAME")

	var cfg testConfig
	c.Assert(Load(&cfg), qt.ErrorMatches, ".*APP_NAME.*")
}
