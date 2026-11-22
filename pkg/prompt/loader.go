package prompt

import (
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// LoadUserTemplates loads all YAML templates from the given directory.
func LoadUserTemplates(dir string) (map[string]*Template, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	templates := make(map[string]*Template)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var tpl Template
			if err := yaml.Unmarshal(data, &tpl); err != nil {
				continue
			}
			key := tpl.Name + "@v" + tpl.Version
			templates[key] = &tpl
		}
	}
	return templates, nil
}