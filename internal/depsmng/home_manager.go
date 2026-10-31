package depsmng

import (
	"embed"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/template"
)

//go:embed templates/flake.nix.tmpl
var templateFS embed.FS

type UserData struct {
	Username string
	HomeDirectory string
}

type FlakeConfig struct {
    Channel            string
    HomeManagerChannel string
    System             string
    Pkgs               map[string]struct{}
    HomeStateVersion   string
    Username           string
    HomeDirectory      string
}

var (
	nixChannel 		   = "github:NixOS/nixpkgs/nixos-26.05"
	homeManagerChannel = "github:nix-community/home-manager/release-26.05"
	allowedSystem      = "x86_64-linux"
	stateVersion       = "26.05"
)

func (fc *FlakeConfig) New(data *UserData) *FlakeConfig {
	return &FlakeConfig{
		Channel: nixChannel,
		HomeManagerChannel: homeManagerChannel,
		System: allowedSystem,
		Pkgs: make(map[string]struct{}),
		HomeStateVersion: stateVersion,
		Username: data.Username,
		HomeDirectory: data.HomeDirectory,
	}	
}

func (fc *FlakeConfig) addDependency(dependency string) {
	if  fc.Pkgs == nil {
		fc.Pkgs = make(map[string]struct{})
	}

	fc.Pkgs[dependency] = struct{}{}
}

func (fc *FlakeConfig) removeDependency(dependency string) {
	if  fc.Pkgs == nil {
		return
	}

	delete(fc.Pkgs, dependency)
}

func (fc *FlakeConfig) switchProfile() error {
	return nil
}

func (fc *FlakeConfig) checkDependency(dependency string) error {
	return nil
}

func (fc *FlakeConfig) SetupConfig() error {
	info, err := os.Stat(fc.flakeDirPath())

	if os.IsNotExist(err) {
		err := os.MkdirAll(fc.flakeDirPath(), 0755)

		if err != nil {
			return err
		}
	} else if info.IsDir() == false {
		return fmt.Errorf("FlakeDir is not a directory")
	}

	tmpl, err := fc.renderFlake()
	if err != nil {
		return err 
	}

	file, err := os.OpenFile(fc.flakeConfigPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, *fc)
}

func (fc *FlakeConfig) renderFlake() (*template.Template, error) {
	tmpl, err := template.
		New("flake.nix.tmpl").
		Funcs(template.FuncMap{
			"join": func(pkgs map[string]struct{}) string {
				names := make([]string, 0, len(pkgs))

				for name := range pkgs {
					names = append(names, name)
				}

				slices.Sort(names)

				return strings.Join(names, ", ")
			},
		}).
		ParseFS(templateFS, "templates/flake.nix.tmpl")

	if err != nil {
		return nil, nil
	}

	return tmpl, nil
}

func (fc *FlakeConfig) flakeDirPath() string {
	return fmt.Sprintf("%s/.kubease", fc.HomeDirectory)
}

func (fc *FlakeConfig) flakeConfigPath() string {
	return fmt.Sprintf("%s/flake.nix", fc.flakeDirPath())
}
