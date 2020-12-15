package project

// FilterOptions contains a list of secret class filter options
type FilterOptions struct {
	DefaultAll bool
	Add        []string
	Subtract   []string
}

func filterSecrets(secrets []SecretConfig, options FilterOptions) ([]SecretConfig, []SecretConfig) {
	selectedSecrets := []SecretConfig{}
	unselectedSecrets := []SecretConfig{}

	for _, secret := range secrets {
		shouldAppend := false

		if secret.Class == nil {
			shouldAppend = true
		} else {
			if options.DefaultAll {
				shouldAppend = true
			} else {
				for _, class := range options.Add {
					if *secret.Class == class {
						shouldAppend = true
						break
					}
				}
			}

			for _, class := range options.Subtract {
				if *secret.Class == class {
					shouldAppend = false
					break
				}
			}
		}

		if shouldAppend {
			selectedSecrets = append(selectedSecrets, secret)
		} else {
			unselectedSecrets = append(unselectedSecrets, secret)
		}
	}

	return selectedSecrets, unselectedSecrets
}
