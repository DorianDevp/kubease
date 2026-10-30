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
	if err := addDependency("kubernetes"); err != nil {
		return err
	}

	if err := switchProfile(); err != nil {
		return err
	}

	if err := checkDep("kubeadm"); err != nil {
		return err
	}

	if err := checkDep("kubectl"); err != nil {
		return err
	}

	if err := checkDep("kubelet"); err != nil {
		return err
	}

	return nil
}
