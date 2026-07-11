// Package registry is the public runtime-registration API through which a
// CONSUMER project injects its project-specific data — service adapters, the
// app roster, and named endpoints — into the otherwise project-agnostic engine.
//
// This is the load-bearing §11.4.28 decoupling seam: the HelixStream engine
// carries NO project literal (no package names, device serials, or region
// endpoints). The consumer (e.g. ATMOSphere) calls RegisterAdapter /
// RegisterRoster / RegisterEndpoint at startup with ITS data. The engine reads
// only what the consumer registered.
//
// P1 NOTE (skeleton): the registry is complete and used; the adapters it will
// hold are consumer data authored in later PWUs (P8/P9/P11).
package registry

import (
	"errors"
	"fmt"
	"sort"

	"digital.vasic.helixstream/pkg/adapter"
)

// RosterEntry describes one installed streaming app the consumer wants tested.
// All fields are consumer data — the engine defines the SHAPE, never the values.
type RosterEntry struct {
	Package       string // e.g. the app's package id (consumer data)
	AppName       string
	Category      string // video | music | mixed (consumer taxonomy)
	GeoRestricted bool
	RequiresLogin bool
}

var (
	// ErrDuplicate is returned when registering a name that already exists.
	ErrDuplicate = errors.New("registry: duplicate registration")
	// ErrEmptyName is returned when an adapter/endpoint name is empty.
	ErrEmptyName = errors.New("registry: empty name")
)

// Registry holds the consumer-registered data for one engine run.
type Registry struct {
	adapters  map[string]adapter.ServiceAdapter
	endpoints map[string]string
	roster    []RosterEntry
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		adapters:  map[string]adapter.ServiceAdapter{},
		endpoints: map[string]string{},
	}
}

// RegisterAdapter registers a consumer-supplied service adapter under its Name().
func (r *Registry) RegisterAdapter(a adapter.ServiceAdapter) error {
	if a == nil {
		return errors.New("registry: nil adapter")
	}
	if a.Name() == "" {
		return ErrEmptyName
	}
	if _, ok := r.adapters[a.Name()]; ok {
		return fmt.Errorf("%w: adapter %q", ErrDuplicate, a.Name())
	}
	r.adapters[a.Name()] = a
	return nil
}

// Adapter returns a registered adapter by name.
func (r *Registry) Adapter(name string) (adapter.ServiceAdapter, bool) {
	a, ok := r.adapters[name]
	return a, ok
}

// Names returns the sorted set of registered adapter names.
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.adapters))
	for n := range r.adapters {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// RegisterRoster appends consumer roster entries.
func (r *Registry) RegisterRoster(entries []RosterEntry) error {
	for _, e := range entries {
		if e.Package == "" {
			return fmt.Errorf("%w: roster entry with empty package", ErrEmptyName)
		}
	}
	r.roster = append(r.roster, entries...)
	return nil
}

// Roster returns the registered roster (a copy).
func (r *Registry) Roster() []RosterEntry {
	return append([]RosterEntry(nil), r.roster...)
}

// RegisterEndpoint registers a named consumer endpoint (e.g. a sink probe URL).
func (r *Registry) RegisterEndpoint(name, url string) error {
	if name == "" {
		return ErrEmptyName
	}
	if _, ok := r.endpoints[name]; ok {
		return fmt.Errorf("%w: endpoint %q", ErrDuplicate, name)
	}
	r.endpoints[name] = url
	return nil
}

// Endpoint returns a registered endpoint URL by name.
func (r *Registry) Endpoint(name string) (string, bool) {
	u, ok := r.endpoints[name]
	return u, ok
}
