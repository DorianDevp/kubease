package depsmng

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strings"
)

//go:embed templates/flake.nix.tmpl
var templateFS embed.FS

type homeManager struct {
	Flake *FlakeConfig
}

func (hm *homeManager) NewFlake(data *user.User) *FlakeConfig {
	fl := &FlakeConfig{
		Channel:            nixChannel,
		HomeManagerChannel: homeManagerChannel,
		System:             allowedSystem,
		Pkgs:               make(map[string]struct{}),
		HomeStateVersion:   stateVersion,
		Username:           data.Username,
		HomeDirectory:      data.HomeDir,
	}

	hm.Flake = fl

	return fl
}

func (hm *homeManager) switchProfile() error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	fc := hm.Flake
	if fc == nil {
		return fmt.Errorf("Empty flake")
	}

	flakeRef := fmt.Sprintf(
		"%s#%s",
		fc.flakeDirPath(),
		fc.Username,
	)

	cmd := exec.Command(
		"home-manager",
		"switch",
		"--flake",
		flakeRef,
	)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"Home Manager switch failed: %w\n%s",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}

	if output := strings.TrimSpace(stdout.String()); output != "" {
		log.Print(output)
	}

	return nil
}

func buildKubeaseNixFlake(hm *homeManager, user *user.User) (*FlakeConfig, error) {
	fk := hm.NewFlake(user)

	fk.addDependency("kubernetes")
	fk.addDependency("wireguard-tools")
	fk.addDependency("go")

	if err := fk.SetupConfig(); err != nil {
		return nil, err
	}

	return fk, nil
}
