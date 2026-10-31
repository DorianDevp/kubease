package depsmng

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultNixConfPath = "/etc/nix/nix.conf"
	nixInstallerURL    = "https://install.determinate.systems/nix"
)

var requiredExperimentalFeatures = []string{
	"nix-command",
	"flakes",
}

type SysdepsSetupError struct {
	Operation string
	Err       error
}

func (e *SysdepsSetupError) Error() string {
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

func installNix() error {
	resp, err := http.Get(nixInstallerURL)
	if err != nil {
		return &SysdepsSetupError{
			Operation: "Nix installer download failed",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &SysdepsSetupError{
			Operation: "Nix installer download failed",
			Err:       fmt.Errorf("unexpected HTTP status: %s", resp.Status),
		}
	}

	installerScript, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SysdepsSetupError{
			Operation: "Nix installer download failed",
			Err:       err,
		}
	}

	cmd := exec.Command("sh", "-c", string(installerScript))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return &SysdepsSetupError{
			Operation: "Nix installation failed",
			Err:       err,
		}
	}

	return nil
}

func checkFlakes() error {
	configContent, err := os.ReadFile(defaultNixConfPath)
	if err != nil {
		return &SysdepsSetupError{
			Operation: "Nix configuration check failed",
			Err:       fmt.Errorf("cannot read %q: %w", defaultNixConfPath, err),
		}
	}

	config := string(configContent)

	if err := hasExperimentalFeatures(config); err != nil {
		return &SysdepsSetupError{
			Operation: "Nix configuration check failed",
			Err:       fmt.Errorf("missing required experimental feature: %s", err),
		}
	}

	return nil
}

func hasExperimentalFeatures(config string) error {
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(key) != "experimental-features" {
			continue
		}

		errAgg := ""
		for _, feature := range requiredExperimentalFeatures {
			found := false

			for _, configuredFeature := range strings.Fields(value) {
				if found == false && configuredFeature == feature {
					found = true
				}
			}

			if found == true {
				errAgg = fmt.Sprintf("%s %s", errAgg, feature)
			}
		}

		if errAgg != "" {
			return fmt.Errorf("%s", strings.TrimSpace(errAgg))
		}
	}

	return nil 
}

