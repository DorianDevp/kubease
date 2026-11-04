package depsmng

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime/debug"
	"testing"
)

func TestBootstrap(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	} else if usr.HomeDir == "" || usr.Username == "" {
		t.Fatalf("HomeDir or Username are empty")
	}

	if _, err := Bootstrap(); err != nil {
		log.Fatalf("Error: %v\n%s", err, debug.Stack())

		t.Fatal(err)
	}

	filesToAssert := []string{
		defaultNixConfPath,
		fmt.Sprintf("%s/.config/home-manager/home.nix", usr.HomeDir),
		fmt.Sprintf("%s/.kubease/flake.nix", usr.HomeDir),
	}

	for _, file := range filesToAssert {
		if _, err = os.Stat(file); err != nil {
			t.Fatal(err)
		}
	}

	executablesToCheck := []string{
		"kubectl",
		"kubeadm",
		"wg",
	}

	for _, executable := range executablesToCheck {
		if _, err := exec.LookPath(executable); err != nil {
			t.Fatal(err)
		}
	}
}
