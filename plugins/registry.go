package plugins

import "fmt"

var registry = map[string]DistroPlugin{}

func Register(p DistroPlugin) {
	registry[p.ID()] = p
}

func Get(id string) (DistroPlugin, error) {
	if p, ok := registry[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("no plugin registered for distro %q", id)
}
