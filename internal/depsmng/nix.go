package depsmng

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	defaultNixConfPath = "/etc/nix/nix.conf"
	nixInstallerURL    = "https://nixos.org/nix/install"
	homeManagerArchive = "https://github.com/nix-community/home-manager/archive/release-26.05.tar.gz"
)

var requiredExperimentalFeatures = []string{
	"nix-command",
	"flakes",
}

type nix struct {
	ConfigPath string
	HomeManager *homeManager
}

type SysdepsSetupError struct {
	Operation string
	Err       error
}

func (e *SysdepsSetupError) Error() string {
	return fmt.Sprintf("%s: %v", e.Operation, e.Err)
}

// 
// SingleUserInstallation: 
// this app uses nix only for **specific** packages cohesion. Those packages
// should be used and managed only by root, so multi-user does not fit our case.
// Also multi-user insist on having systemd (by service setup) which makes testing
// in containers a bit harder
//
func installNix() (*nix, error) {
	// @FIXME: download source, and install on runtime instead of http request
    resp, err := http.Get(nixInstallerURL)
    if err != nil {
        return nil, &SysdepsSetupError{
            Operation: "Nix installer download failed",
            Err:       err,
        }
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, &SysdepsSetupError{
            Operation: "Nix installer download failed",
            Err:       fmt.Errorf("unexpected HTTP status: %s", resp.Status),
        }
    }

    if err := os.MkdirAll("/nix", 0755); err != nil {
        return nil, err
    }

    cmd := exec.Command(
        "bash",
        "-s",
        "--",
        "--no-daemon",
    )

    cmd.Stdin = resp.Body
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    cmd.Env = append(os.Environ(),
        "NIX_CONFIG=build-users-group =\nsandbox = false\nexperimental-features = nix-command flakes",
    )

    if err := cmd.Run(); err != nil {
        return nil, &SysdepsSetupError{
            Operation: "Nix installation failed",
            Err:       err,
        }
    }

	n := &nix{
		ConfigPath: defaultNixConfPath,
	}

	if err := n.configureNix(); err != nil {
		return nil, err
	}

	if err := loadNixEnvironment(); err != nil {
		return nil, err
	}

	return n, nil
}

func (n *nix) configureNix() error {
    if err := os.MkdirAll(filepath.Dir(n.ConfigPath), 0755); err != nil {
        return err
    }

    config := strings.Join([]string{
        "build-users-group =",
        "sandbox = false",
        "experimental-features = nix-command flakes",
    }, "\n") + "\n"

    if err := os.WriteFile(n.ConfigPath, []byte(config), 0644); err != nil {
        return err
    }

    return nil
}

func (n *nix) installHomeManager() error {
	if _, err := exec.LookPath("home-manager"); err == nil {
		return nil
	} 

	fmt.Println("PATH:", os.Getenv("PATH"))

	path, err := exec.LookPath("nix-channel")
	fmt.Println("nix-channel:", path, err)
	if _, err := exec.LookPath("nix-channel"); err != nil {
		return err
	}

	run := func (name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s %v failed: %w", name, args, err)
		}

		return nil
	}

    if err := run(
        "nix-channel",
        "--add",
        homeManagerArchive,
        "home-manager",
    ); err != nil {
        return err
    }

    if err := run("nix-channel", "--update"); err != nil {
        return err
    }

	if err := run(
        "nix-shell",
        "<home-manager>",
        "-A",
        "install",
    ); err != nil {
		return err
	}

	n.HomeManager = &homeManager{}

	return nil
}

func (n *nix) assertFlakesRequirements() error {

	return nil
}

func (n *nix) hasExperimentalFeatures() (bool, error) {
	configContent, err := os.ReadFile(n.ConfigPath)
	if err != nil {
		return false, &SysdepsSetupError{
			Operation: "Nix configuration check failed",
			Err:       fmt.Errorf("cannot read %q: %w", defaultNixConfPath, err),
		}
	}

	config := string(configContent)

    enabled := make(map[string]struct{})

    for _, line := range strings.Split(config, "\n") {
        line = strings.TrimSpace(line)

        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        line, _, _ = strings.Cut(line, "#")

        key, value, found := strings.Cut(line, "=")
        if !found || strings.TrimSpace(key) != "experimental-features" {
            continue
        }

        for _, feature := range strings.Fields(value) {
            enabled[feature] = struct{}{}
        }
    }

    var missing []string

    for _, feature := range requiredExperimentalFeatures {
        if _, found := enabled[feature]; !found {
            missing = append(missing, feature)
        }
    }

    if len(missing) == 0 {
		return true, nil
    }

	return false, &SysdepsSetupError{
		Operation: "Nix configuration check failed",
		Err:       fmt.Errorf("%s", strings.Join(missing, ", ")),
	}
}

func loadNixEnvironment() error {
    currentUser, err := user.Current()
    if err != nil {
        return fmt.Errorf("getting current user: %w", err)
    }

    if err := os.Setenv("USER", currentUser.Username); err != nil {
        return err
    }

    profileBin := filepath.Join(
        currentUser.HomeDir,
        ".nix-profile",
        "bin",
    )

    path := os.Getenv("PATH")

    if err := os.Setenv(
        "PATH",
        profileBin+string(os.PathListSeparator)+path,
    ); err != nil {
        return err
    }

    return nil
}
