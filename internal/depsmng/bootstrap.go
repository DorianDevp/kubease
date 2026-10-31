package depsmng

func Bootstrap() error {
	if err := installNix(); err != nil {
		return err
	}

	if err := checkFlakes(); err != nil {
		return err
	}

	if err := buildKubernetes(); err != nil {
		return err
	}

	
	return nil
}

func buildKubernetes() error {

	return nil
}
