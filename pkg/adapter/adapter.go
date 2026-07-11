// Package adapter defines the per-service ServiceAdapter contract (DESIGN.md
// §3.2 NEW, §10): a streaming service is described to the engine by composing
// the three sub-contracts (login state machine + catalog map + playback
// controller) plus service metadata (kind, required country).
//
// The engine provides the generic FSM/catalog/playback DRIVERS; an adapter only
// supplies anchors + capability declarations. Concrete adapters (START, IVI,
// Wink, OKKO, VK, Zvuk, …) are consumer-owned DATA registered at runtime via
// pkg/registry (§11.4.28 decoupling) and are built in later PWUs (P8/P9).
//
// P1 NOTE (skeleton): the contract + a Stub composed from the sub-package stubs.
package adapter

import (
	"digital.vasic.helixstream/pkg/catalog"
	"digital.vasic.helixstream/pkg/loginsm"
	"digital.vasic.helixstream/pkg/playback"
)

// Kind distinguishes a video service (full video+subtitle+2nd-display path)
// from a music service (audio-only path).
type Kind string

const (
	Video Kind = "video"
	Music Kind = "music"
)

// CountryAny is the sentinel meaning "no geo requirement" for RequiredCountry.
const CountryAny = "any"

// ServiceAdapter is the core per-service contract. Everything a streaming
// service needs to be driven is reachable through it.
type ServiceAdapter interface {
	// Name is the stable adapter name (e.g. the app's short id). It is consumer
	// data — the engine imposes no naming.
	Name() string
	// Kind reports video vs music.
	Kind() Kind
	// RequiredCountry is an ISO country code the service needs, or CountryAny.
	// Consumed by the geo-gate (§8.2); the engine never hardcodes a country.
	RequiredCountry() string
	// Login returns the login state machine for this service.
	Login() loginsm.LoginStateMachine
	// Catalog returns the catalog-browse map for this service.
	Catalog() catalog.CatalogMap
	// Playback returns the playback controller for this service.
	Playback() playback.PlaybackController
}

// Stub is a minimal composed adapter used for contract tests and as the base
// an in-tree adapter can embed. It carries NO project-specific literal.
type Stub struct {
	name    string
	kind    Kind
	country string
	login   loginsm.LoginStateMachine
	catalog catalog.CatalogMap
	play    playback.PlaybackController
}

// NewStub builds a Stub adapter. Any nil sub-contract is replaced with the
// corresponding package's own stub so the result is always fully usable.
func NewStub(name string, kind Kind, country string,
	login loginsm.LoginStateMachine, cat catalog.CatalogMap, play playback.PlaybackController) *Stub {
	if login == nil {
		login = loginsm.NewStub()
	}
	if cat == nil {
		cat = catalog.NewStub(nil, nil)
	}
	if play == nil {
		play = playback.NewStub(nil)
	}
	if country == "" {
		country = CountryAny
	}
	return &Stub{name: name, kind: kind, country: country, login: login, catalog: cat, play: play}
}

func (s *Stub) Name() string                          { return s.name }
func (s *Stub) Kind() Kind                            { return s.kind }
func (s *Stub) RequiredCountry() string               { return s.country }
func (s *Stub) Login() loginsm.LoginStateMachine      { return s.login }
func (s *Stub) Catalog() catalog.CatalogMap           { return s.catalog }
func (s *Stub) Playback() playback.PlaybackController { return s.play }

// Compile-time assertion that Stub satisfies ServiceAdapter.
var _ ServiceAdapter = (*Stub)(nil)
