package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// writeConfigFile creates a config file (and its parent directories).
func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// isolateConfig points HOME and XDG_CONFIG_HOME at temp dirs so config
// discovery in tests never sees the developer's real files, and resets the
// --config flag.
func isolateConfig(t *testing.T) (home, xdg string) {
	t.Helper()
	home = t.TempDir()
	xdg = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	origCfgFile := cfgFile
	t.Cleanup(func() { cfgFile = origCfgFile })
	cfgFile = ""
	return home, xdg
}

func xdgConfigPath(xdg string) string {
	return filepath.Join(xdg, "git-open", "config.yaml")
}

func legacyConfigPath(home string) string {
	return filepath.Join(home, ".git-open.yaml")
}

func Test_configFilePath(t *testing.T) {
	t.Run("--config overrides the defaults", func(t *testing.T) {
		home, xdg := isolateConfig(t)
		writeConfigFile(t, xdgConfigPath(xdg), "browser: xdg\n")
		writeConfigFile(t, legacyConfigPath(home), "browser: legacy\n")
		custom := filepath.Join(t.TempDir(), "custom.yaml")
		writeConfigFile(t, custom, "browser: custom\n")
		cfgFile = custom

		if got := configFilePath(); got != custom {
			t.Errorf("configFilePath() = %q, want %q", got, custom)
		}
	})

	t.Run("XDG config preferred over legacy", func(t *testing.T) {
		home, xdg := isolateConfig(t)
		writeConfigFile(t, xdgConfigPath(xdg), "browser: xdg\n")
		writeConfigFile(t, legacyConfigPath(home), "browser: legacy\n")

		if got, want := configFilePath(), xdgConfigPath(xdg); got != want {
			t.Errorf("configFilePath() = %q, want %q", got, want)
		}
	})

	t.Run("legacy config as fallback", func(t *testing.T) {
		home, _ := isolateConfig(t)
		writeConfigFile(t, legacyConfigPath(home), "browser: legacy\n")

		if got, want := configFilePath(), legacyConfigPath(home); got != want {
			t.Errorf("configFilePath() = %q, want %q", got, want)
		}
	})

	t.Run("no config file at all", func(t *testing.T) {
		isolateConfig(t)
		if got := configFilePath(); got != "" {
			t.Errorf("configFilePath() = %q, want empty", got)
		}
	})

	t.Run("~/.config used when XDG_CONFIG_HOME is unset", func(t *testing.T) {
		home, _ := isolateConfig(t)
		t.Setenv("XDG_CONFIG_HOME", "")
		want := filepath.Join(home, ".config", "git-open", "config.yaml")
		writeConfigFile(t, want, "browser: dot-config\n")

		if got := configFilePath(); got != want {
			t.Errorf("configFilePath() = %q, want %q", got, want)
		}
	})

	t.Run("relative XDG_CONFIG_HOME is ignored", func(t *testing.T) {
		home, _ := isolateConfig(t)
		t.Setenv("XDG_CONFIG_HOME", "relative/config")
		want := filepath.Join(home, ".config", "git-open", "config.yaml")
		writeConfigFile(t, want, "browser: dot-config\n")

		if got := configFilePath(); got != want {
			t.Errorf("configFilePath() = %q, want %q", got, want)
		}
	})

	t.Run("directories are not treated as config files", func(t *testing.T) {
		home, xdg := isolateConfig(t)
		if err := os.MkdirAll(xdgConfigPath(xdg), 0o755); err != nil {
			t.Fatal(err)
		}
		writeConfigFile(t, legacyConfigPath(home), "browser: legacy\n")

		if got, want := configFilePath(), legacyConfigPath(home); got != want {
			t.Errorf("configFilePath() = %q, want %q", got, want)
		}
	})
}

// Test_initConfigXDG checks that the XDG config file is loaded and wins
// over the legacy ~/.git-open.yaml.
func Test_initConfigXDG(t *testing.T) {
	originalBrowserCommand := BrowserCommand
	t.Cleanup(func() { BrowserCommand = originalBrowserCommand })
	t.Setenv("BROWSER", "")

	home, xdg := isolateConfig(t)
	writeConfigFile(t, xdgConfigPath(xdg), "browser: xdg-browser\n")
	writeConfigFile(t, legacyConfigPath(home), "browser: legacy-browser\n")

	initConfig()
	if BrowserCommand != "xdg-browser" {
		t.Errorf("BrowserCommand = %q, want %q (XDG config must win)", BrowserCommand, "xdg-browser")
	}
}

// Test_initConfigLegacyFallback checks that the legacy location still
// loads when no XDG config exists.
func Test_initConfigLegacyFallback(t *testing.T) {
	originalBrowserCommand := BrowserCommand
	t.Cleanup(func() { BrowserCommand = originalBrowserCommand })
	t.Setenv("BROWSER", "")

	home, _ := isolateConfig(t)
	writeConfigFile(t, legacyConfigPath(home), "browser: legacy-browser\n")

	initConfig()
	if BrowserCommand != "legacy-browser" {
		t.Errorf("BrowserCommand = %q, want %q", BrowserCommand, "legacy-browser")
	}
}

// Test_initConfigNoConfig checks that a missing config file is not an
// error and leaves the browser unset (environment only).
func Test_initConfigNoConfig(t *testing.T) {
	originalBrowserCommand := BrowserCommand
	t.Cleanup(func() { BrowserCommand = originalBrowserCommand })
	t.Setenv("BROWSER", "")

	isolateConfig(t)

	initConfig()
	if BrowserCommand != "" {
		t.Errorf("BrowserCommand = %q, want empty", BrowserCommand)
	}
}

// Test_xdgConfigHomeWithoutHome covers the fallback when neither
// XDG_CONFIG_HOME nor HOME can provide a directory.
func Test_xdgConfigHomeWithoutHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")

	if got := xdgConfigHome(); got != "" {
		t.Errorf("xdgConfigHome() = %q, want empty without HOME", got)
	}
}
