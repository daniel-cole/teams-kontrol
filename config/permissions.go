package config

type Permissions struct {
	Verbs      []string `yaml:"verbs"`
	Resources  []string `yaml:"resources"`
	Namespaces []string `yaml:"namespaces"`
}
