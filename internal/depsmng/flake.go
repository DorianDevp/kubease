package depsmng

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

const (
	nixChannel          = "github:NixOS/nixpkgs/nixos-26.05"
	homeManagerChannel  = "github:nix-community/home-manager/release-26.05"
	allowedSystem       = "x86_64-linux"
	stateVersion        = "26.05"
)

type FlakeConfig struct {
	Channel            string
	HomeManagerChannel string
	System             string
	Pkgs               map[string]struct{}
	HomeStateVersion   string
	Username           string
	HomeDirectory      string
}

func (fc *FlakeConfig) addDependency(dependency string) {
	if fc.Pkgs == nil {
		fc.Pkgs = make(map[string]struct{})
	}

	fc.Pkgs[dependency] = struct{}{}
}

func (fc *FlakeConfig) removeDependency(dependency string) {
	if fc.Pkgs == nil {
		return
	}

	delete(fc.Pkgs, dependency)
}

func (fc *FlakeConfig) SetupConfig() error {
	if err := os.MkdirAll(fc.flakeDirPath(), 0755); err != nil {
		return err
	}

	tmpl, err := fc.renderFlake()
	if err != nil {
		return err
	}

	var content bytes.Buffer

	if err := tmpl.Execute(&content, fc); err != nil {
		return err
	}

	return os.WriteFile(fc.flakeConfigPath(), content.Bytes(), 0644)
}

func (fc *FlakeConfig) renderFlake() (*template.Template, error) {
	return template.
		New("flake.nix.tmpl").
		Funcs(template.FuncMap{
			"join": func(pkgs map[string]struct{}) string {
				names := make([]string, 0, len(pkgs))

				for name := range pkgs {
					names = append(names, name)
				}

				slices.Sort(names)

				return strings.Join(names, "\n")
			},
			"quote": strconv.Quote,
		}).
		ParseFS(templateFS, "templates/flake.nix.tmpl")
}

func (fc *FlakeConfig) flakeDirPath() string {
	return filepath.Join(fc.HomeDirectory, ".kubease")
}

func (fc *FlakeConfig) flakeConfigPath() string {
	return filepath.Join(fc.flakeDirPath(), "flake.nix")
}
