package prompt

import (
	"errors"
	"fmt"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
)

// Option is a functional option for Registry configuration.
type Option func(*Registry)

// WithUserDir sets the user template directory.
func WithUserDir(dir string) Option {
	return func(r *Registry) {
		r.userDir = dir
	}
}

// Registry holds built-in and user templates.
type Registry struct {
	builtins map[string]*Template
	userDir  string
	userTpls map[string]*Template
}

// NewRegistry creates a new Registry with built-in templates.
func NewRegistry(opts ...Option) *Registry {
	r := &Registry{
		builtins: builtinTemplates(),
		userTpls: make(map[string]*Template),
	}
	for _, opt := range opts {
		opt(r)
	}
	// Load user templates if userDir is set
	if r.userDir != "" {
		tpls, _ := LoadUserTemplates(r.userDir)
		r.userTpls = tpls
	}
	return r
}

// Get retrieves a template by name and version.
// Checks userDir first, then builtins.
func (r *Registry) Get(name, version string) (*Template, error) {
	key := name + "@v" + version

	// Check user templates first
	if tpl, ok := r.userTpls[key]; ok {
		return tpl, nil
	}

	// Check builtins
	if tpl, ok := r.builtins[key]; ok {
		return tpl, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, key)
}

// List returns all available template keys (name@version).
func (r *Registry) List() []string {
	keys := make([]string, 0, len(r.builtins)+len(r.userTpls))
	for k := range r.builtins {
		keys = append(keys, k)
	}
	for k := range r.userTpls {
		keys = append(keys, k)
	}
	return keys
}