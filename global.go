package naminggrpc

import (
	"strings"
	"sync"
)

var registrarMu sync.RWMutex

var registrars = make(map[string]Factory)

func Register(name string, resource Factory) {
	if resource == nil {
		panic("gonfig: RegisterResource resource is nil")
	}
	name = strings.ToLower(name)
	registrarMu.Lock()
	defer registrarMu.Unlock()
	if _, dup := registrars[name]; dup {
		panic("gonfig: RegisterResource called twice for resource " + name)
	}
	registrars[name] = resource
}

func Get(name string) (Factory, bool) {
	name = strings.ToLower(name)
	registrarMu.RLock()
	defer registrarMu.RUnlock()
	resource, ok := registrars[name]
	return resource, ok
}
