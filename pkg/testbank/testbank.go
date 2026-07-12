// Package testbank defines the VERSIONED streaming-app test-bank schema and a
// fully generic, project-agnostic loader (DESIGN.md §5).
//
// A "test bank" is consumer-owned YAML DATA describing which streaming services
// to test and how (per-service adapter ref, catalog probe, playback-
// manipulation steps, smooth-AV thresholds, topology + geo requirements). The
// operator mandate is that banks are VERSION-TRACKED: both the bank-schema
// version and the bank's own version are mandatory, and a bank missing either
// is REJECTED at load — never silently accepted (§11.4.6 / §11.4.107(10)).
//
// §11.4.27 / §11.4.28 — the LOADER MACHINERY here is 100% generic: it carries
// ZERO project literal (no service names, package ids, device serials, or
// region endpoints). All of that lives in the consumer's YAML. The P1
// decoupling audit (internal/decoupling) scans this package and MUST continue
// to find zero project literals.
//
// The schema ties to the P1 contracts: ServiceEntry.Kind maps to adapter.Kind,
// Topology maps to topology.Topology, Geo.RequiredCountry is consumed by
// pkg/geogate, Playback.SmoothAV maps to smoothav.Thresholds.
package testbank

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"

	"digital.vasic.helixstream/pkg/smoothav"
)

// SupportedSchemaVersions is the closed set of bank-schema versions this loader
// understands. A bank declaring any other schema_version is REJECTED.
var SupportedSchemaVersions = map[string]bool{"1.0": true}

// Bank is a versioned streaming-app test bank (project-agnostic schema).
type Bank struct {
	SchemaVersion string         `yaml:"schema_version"` // bank-SCHEMA version the loader validates (mandatory)
	Name          string         `yaml:"name"`
	BankVersion   string         `yaml:"bank_version"` // the version-tracked bank version (mandatory)
	Services      []ServiceEntry `yaml:"services"`
}

// ServiceEntry describes one streaming service to test. id/adapter/kind/topology
// are validated strictly; the rest is consumer data.
type ServiceEntry struct {
	ID               string        `yaml:"id"`               // stable service id (consumer data)
	Name             string        `yaml:"name"`             // human label (optional)
	Package          string        `yaml:"package"`          // consumer app package id (optional, consumer data)
	Adapter          string        `yaml:"adapter"`          // adapter ref (a pkg/registry-registered adapter name)
	MinAppVersion    string        `yaml:"min_app_version"`  // version pin (optional)
	Kind             string        `yaml:"kind"`             // video | music (maps to adapter.Kind)
	Topology         string        `yaml:"topology"`         // any | dual_display_only | single_display_only
	Geo              GeoReq        `yaml:"geo"`
	Catalog          CatalogProbe  `yaml:"catalog"`
	Playback         PlaybackSteps `yaml:"playback"`
	RequiredEvidence []string      `yaml:"required_evidence"`
}

// GeoReq is the per-service geo requirement (consumed by pkg/geogate, §8.2).
type GeoReq struct {
	RequiredCountry string `yaml:"required_country"` // "" / any / US / NO / RU / 2-letter ISO
	NativeFirst     bool   `yaml:"native_first"`
}

// CatalogProbe is the catalog-browse probe for a service (pkg/catalog).
type CatalogProbe struct {
	ProbeQuery string `yaml:"probe_query"`
}

// PlaybackSteps is the playback-manipulation spec for a service (pkg/playback).
type PlaybackSteps struct {
	HighestVideo        bool        `yaml:"highest_video"`
	HighestAudio        bool        `yaml:"highest_audio"`
	RandomSeeks         int         `yaml:"random_seeks"`
	Transport           []string    `yaml:"transport"` // subset of play|pause|resume|stop
	SelectAudioTrack    bool        `yaml:"select_audio_track"`
	SelectSubtitleTrack bool        `yaml:"select_subtitle_track"`
	SmoothAV            SmoothAVReq `yaml:"smooth_av"`
}

// SmoothAVReq maps to smoothav.Thresholds — the smooth-AV acceptance budgets
// (§8.1). ToThresholds bridges the bank data to the P1 smoothav contract.
type SmoothAVReq struct {
	FPSTolerance float64 `yaml:"fps_tolerance"`
	MaxDropped   int     `yaml:"max_dropped"`
	FreezeSSIM   float64 `yaml:"freeze_ssim"`
	MaxJitterMS  float64 `yaml:"max_jitter_ms"`
	MinRMSFloor  float64 `yaml:"min_rms_floor"`
	MaxXRUNBurst int     `yaml:"max_xrun_burst"`
}

// ToThresholds converts the bank's smooth-AV budgets to a smoothav.Thresholds,
// tying the schema to the P1 acceptance contract.
func (s SmoothAVReq) ToThresholds() smoothav.Thresholds {
	return smoothav.Thresholds{
		FPSTolerance: s.FPSTolerance,
		MaxDropped:   s.MaxDropped,
		FreezeSSIM:   s.FreezeSSIM,
		MaxJitterMS:  s.MaxJitterMS,
		MinRMSFloor:  s.MinRMSFloor,
		MaxXRUNBurst: s.MaxXRUNBurst,
	}
}

// Sentinel validation errors (so tests can assert precisely which rule fired).
var (
	ErrMissingSchemaVersion     = errors.New("testbank: missing schema_version (banks are version-tracked)")
	ErrUnsupportedSchemaVersion = errors.New("testbank: unsupported schema_version")
	ErrMissingName              = errors.New("testbank: missing bank name")
	ErrMissingBankVersion       = errors.New("testbank: missing bank_version (banks are version-tracked)")
	ErrNoServices               = errors.New("testbank: bank has no services")
	ErrMissingServiceID         = errors.New("testbank: service missing id")
	ErrMissingAdapter           = errors.New("testbank: service missing adapter ref")
	ErrInvalidKind              = errors.New("testbank: service kind must be video|music")
	ErrInvalidTopology          = errors.New("testbank: service topology must be any|dual_display_only|single_display_only")
	ErrInvalidGeoCountry        = errors.New("testbank: geo required_country must be empty|any|<2-letter ISO>")
	ErrInvalidTransport         = errors.New("testbank: transport verb must be play|pause|resume|stop")
	ErrNegativeSeeks            = errors.New("testbank: random_seeks must be >= 0")
	ErrNegativeThreshold        = errors.New("testbank: smooth_av thresholds must be >= 0")
)

var (
	validKinds      = map[string]bool{"video": true, "music": true}
	validTopologies = map[string]bool{"any": true, "dual_display_only": true, "single_display_only": true}
	validTransport  = map[string]bool{"play": true, "pause": true, "resume": true, "stop": true}
	isoCountry      = regexp.MustCompile(`^[A-Z]{2}$`)
)

// Validate checks the bank against the schema rules and returns a joined error
// listing every violation (or nil if the bank is valid). It is the anti-bluff
// gate: a malformed or version-less bank is REJECTED here.
func (b *Bank) Validate() error {
	var errs []error

	if b.SchemaVersion == "" {
		errs = append(errs, ErrMissingSchemaVersion)
	} else if !SupportedSchemaVersions[b.SchemaVersion] {
		errs = append(errs, fmt.Errorf("%w: %q", ErrUnsupportedSchemaVersion, b.SchemaVersion))
	}
	if b.Name == "" {
		errs = append(errs, ErrMissingName)
	}
	if b.BankVersion == "" {
		errs = append(errs, ErrMissingBankVersion)
	}
	if len(b.Services) == 0 {
		errs = append(errs, ErrNoServices)
	}

	for i, s := range b.Services {
		where := fmt.Sprintf("service[%d](%q)", i, s.ID)
		if s.ID == "" {
			errs = append(errs, fmt.Errorf("%w: %s", ErrMissingServiceID, where))
		}
		if s.Adapter == "" {
			errs = append(errs, fmt.Errorf("%w: %s", ErrMissingAdapter, where))
		}
		if !validKinds[s.Kind] {
			errs = append(errs, fmt.Errorf("%w: %s got %q", ErrInvalidKind, where, s.Kind))
		}
		if !validTopologies[s.Topology] {
			errs = append(errs, fmt.Errorf("%w: %s got %q", ErrInvalidTopology, where, s.Topology))
		}
		if c := s.Geo.RequiredCountry; c != "" && c != "any" && !isoCountry.MatchString(c) {
			errs = append(errs, fmt.Errorf("%w: %s got %q", ErrInvalidGeoCountry, where, c))
		}
		for _, v := range s.Playback.Transport {
			if !validTransport[v] {
				errs = append(errs, fmt.Errorf("%w: %s got %q", ErrInvalidTransport, where, v))
			}
		}
		if s.Playback.RandomSeeks < 0 {
			errs = append(errs, fmt.Errorf("%w: %s", ErrNegativeSeeks, where))
		}
		av := s.Playback.SmoothAV
		if av.FPSTolerance < 0 || av.FreezeSSIM < 0 || av.MaxJitterMS < 0 || av.MinRMSFloor < 0 ||
			av.MaxDropped < 0 || av.MaxXRUNBurst < 0 {
			errs = append(errs, fmt.Errorf("%w: %s", ErrNegativeThreshold, where))
		}
	}

	return errors.Join(errs...)
}

// Load parses a bank from YAML with STRICT field checking (unknown/typo fields
// are rejected) and then validates it. A malformed, version-less, or
// unknown-field bank returns a non-nil error — never a partially-accepted bank.
func Load(r io.Reader) (*Bank, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true) // reject unknown/typo fields — strict schema
	var b Bank
	if err := dec.Decode(&b); err != nil {
		return nil, fmt.Errorf("testbank: parse: %w", err)
	}
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("testbank: invalid bank: %w", err)
	}
	return &b, nil
}

// LoadFile loads and validates a bank from a file path.
func LoadFile(path string) (*Bank, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("testbank: open %s: %w", path, err)
	}
	defer f.Close()
	return Load(f)
}
