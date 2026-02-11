package k8s

import "gopkg.in/yaml.v2"

func ToYaml(obj any) string {
	yml, _ := yaml.Marshal(obj)
	return string(yml)
}
