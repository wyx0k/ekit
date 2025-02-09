package app

type EmptyConfigLoader struct {
}

func withEmptyConfigLoader() ConfigLoader {
	return &EmptyConfigLoader{}
}
func (e *EmptyConfigLoader) Load(updater *ConfigUpdater) error {
	updater.UpdateConfig(&Conf{})
	return nil
}
