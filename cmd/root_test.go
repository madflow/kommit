package cmd

import (
	"errors"
	"testing"

	"github.com/madflow/kommit/internal/config"
)

func TestInitializeConfigReturnsWrappedErrorForExplicitConfigFile(t *testing.T) {
	oldInit := configInit
	oldGet := configGet
	oldCfgFile := cfgFile

	t.Cleanup(func() {
		configInit = oldInit
		configGet = oldGet
		cfgFile = oldCfgFile
	})

	cfgFile = "/tmp/kommit.yaml"
	configInit = func(path string) error {
		if path != cfgFile {
			t.Fatalf("expected cfgFile %q, got %q", cfgFile, path)
		}
		return errors.New("bad config")
	}
	configGet = func() *config.ResolvedSettings {
		t.Fatal("configGet should not be called on init error")
		return nil
	}

	err := initializeConfig()
	if err == nil {
		t.Fatal("expected initializeConfig error")
	}
	if got := err.Error(); got != "failed to initialize config from /tmp/kommit.yaml: bad config" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestInitializeConfigReturnsWrappedErrorForDefaultLookup(t *testing.T) {
	oldInit := configInit
	oldGet := configGet
	oldCfgFile := cfgFile

	t.Cleanup(func() {
		configInit = oldInit
		configGet = oldGet
		cfgFile = oldCfgFile
	})

	cfgFile = ""
	configInit = func(path string) error {
		if path != "" {
			t.Fatalf("expected empty cfgFile, got %q", path)
		}
		return errors.New("missing file")
	}
	configGet = func() *config.ResolvedSettings {
		t.Fatal("configGet should not be called on init error")
		return nil
	}

	err := initializeConfig()
	if err == nil {
		t.Fatal("expected initializeConfig error")
	}
	if got := err.Error(); got != "failed to initialize config: missing file" {
		t.Fatalf("unexpected error: %q", got)
	}
}
