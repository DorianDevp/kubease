package depsmng

import (
	"os"
	"testing"
)

func TestAddingPackages(t *testing.T) {
	data := &UserData{ Username: "dorian", HomeDirectory: "/home/dorian" }

	var fc FlakeConfig
	flakeConfig := fc.New(data)

	flakeConfig.addDependency("kubernetes")
	flakeConfig.addDependency("wireguard")

	pkgs := flakeConfig.Pkgs

	if _, ok := pkgs["kubernetes"]; ok == false {
		t.Errorf("Expected kubernetes in pkgs")
	}

	if _, ok := pkgs["wireguard"]; ok == false {
		t.Errorf("No wireguard")
	}
}

func TestCreatingConfig(t *testing.T) {
	data := &UserData{ Username: "dorian", HomeDirectory: "/home/dorian" }

	var fc FlakeConfig
	flakeConfig := fc.New(data)

	flakeConfig.addDependency("kubernetes")
	flakeConfig.addDependency("wireguard")

	err := flakeConfig.SetupConfig()
	if err != nil {
		t.Error(err)
	}

	if dir, err := os.Stat(flakeConfig.flakeDirPath()); os.IsNotExist(err) {
		t.Error(err)
		return
	} else if dir.IsDir() == false {
		t.Errorf("Flake Dir Path '%s' was created but is not a directory", flakeConfig.flakeDirPath())
		return
	}

	if _, err := os.Stat(flakeConfig.flakeConfigPath()); os.IsNotExist(err) {
		t.Error(err)
		return
	} 

	// os.RemoveAll(flakeConfig.flakeDirPath())
}
