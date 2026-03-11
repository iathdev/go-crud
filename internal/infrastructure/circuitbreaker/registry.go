package circuitbreaker

import "fmt"

type Registry struct {
	breakers map[string]*Breaker
}

func NewRegistry() *Registry {
	return &Registry{
		breakers: make(map[string]*Breaker),
	}
}

func (registry *Registry) Register(b *Breaker) {
	registry.breakers[b.Name()] = b
}

func (registry *Registry) Get(name string) *Breaker {
	b, ok := registry.breakers[name]
	if !ok {
		panic(fmt.Sprintf("circuit breaker %q not registered", name))
	}
	return b
}
