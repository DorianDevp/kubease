package depsmng

import (
	"os/user"
)

func Bootstrap() (*FlakeConfig, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	nix, err := installNix() 
	if err != nil {
		return nil, err
	}

	if err = nix.installHomeManager(); err != nil {
		return nil, err
	}

	buildKubeaseNixFlake(nix.HomeManager, usr)
	
	if err := nix.HomeManager.switchProfile(); err != nil {
		return nil, err 
	}
	
	return nix.HomeManager.Flake, nil
}

