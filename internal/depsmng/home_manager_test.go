package depsmng

import (
	"os"
	"os/user"
	"testing"
)

func TestAddingPackages(t *testing.T) {
	data := &user.User{ Username: "dorian", HomeDir: "/home/dorian" }

	var hm homeManager
	flakeConfig := hm.NewFlake(data)

	flakeConfig.addDependency("kubernetes")
	flakeConfig.addDependency("wireguard")

	pkgs := flakeConfig.Pkgs

	if _, ok := pkgs["kubernetes"]; ok == false {
		t.Fatalf("Expected kubernetes in pkgs")
	}

	if _, ok := pkgs["wireguard"]; ok == false {
		t.Fatalf("No wireguard")
	}
}

func TestCreatingConfig(t *testing.T) {
	data := &user.User{ Username: "dorian", HomeDir: t.TempDir() }

	var hm homeManager
	flakeConfig := hm.NewFlake(data)

	flakeConfig.addDependency("kubernetes")
	flakeConfig.addDependency("wireguard")

	err := flakeConfig.SetupConfig()
	if err != nil {
		t.Fatal(err)
	}

	if dir, err := os.Stat(flakeConfig.flakeDirPath()); err != nil {
		t.Fatal(err)
	} else if dir.IsDir() == false {
		t.Fatalf("Flake Dir Path '%s' was created but is not a directory", flakeConfig.flakeDirPath())
	}

	if _, err := os.Stat(flakeConfig.flakeConfigPath()); err != nil {
		t.Fatal(err)
	} 
}
