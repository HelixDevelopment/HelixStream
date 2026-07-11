// Package geogate defines the per-device geolocation + VPN gate contract
// (DESIGN.md §8.2): detect the device's outbound-IP country, probe the service
// endpoint's reachability FIRST from the current geo, and only connect a VPN
// (Mullvad) to the required country when the service is genuinely blocked.
//
// A geo-restricted service SKIPs-with-reason (`geo_restricted` /
// `network_unreachable_external`) — it is NEVER a fake PASS and NEVER a
// FAIL-for-geo (§11.4.3 / §11.4.69).
//
// P1 NOTE (skeleton): the contract + a deterministic stub. The real on-device
// geo probe (`adb`), reachability HEAD, and the Mullvad backends (A/C) land in
// P3b; the kernel-WireGuard feasibility is UNCONFIRMED until verified on-device.
package geogate

import "context"

// Reason values (closed set, §11.4.69) used when a service cannot be reached.
const (
	ReasonGeoRestricted  = "geo_restricted"
	ReasonUnreachableExt = "network_unreachable_external"
)

// Outcome is the geo-gate verdict for one service.
type Outcome string

const (
	// ProceedNative — reachable from the current (native) geo, no VPN used.
	ProceedNative Outcome = "proceed_native"
	// ProceedVPN — reachable only after connecting the VPN to the required country.
	ProceedVPN Outcome = "proceed_vpn"
	// SkipRestricted — not reachable even after (or without) VPN → SKIP-with-reason.
	SkipRestricted Outcome = "skip_restricted"
)

// Decision is the captured result of running the gate for one service.
type Decision struct {
	Country    string // detected device country
	Reachable  bool   // reachable at the end of the flow
	UsedVPN    bool   // whether a VPN hop was used
	VPNCountry string // the country the VPN connected to (if UsedVPN)
	Outcome    Outcome
	Reason     string // set when Outcome == SkipRestricted
}

// GeoGate is the core geo/VPN gate contract.
type GeoGate interface {
	// DetectCountry returns the device's current outbound-IP country (ISO).
	DetectCountry(ctx context.Context) (string, error)
	// Reachable probes the service endpoint from the current geo (§11.4.13
	// sink-side, reachability FIRST).
	Reachable(ctx context.Context, serviceEndpoint string) (bool, error)
	// Ensure runs the full reachability-first flow and returns a Decision.
	// requiredCountry is the service's needed country (or "any"). It never
	// FAILs for geo — a genuinely blocked service yields Outcome SkipRestricted.
	Ensure(ctx context.Context, serviceEndpoint, requiredCountry string) (Decision, error)
}

// vpnSupported reports whether the required country is one HelixStream can VPN
// to (DESIGN.md §8.2 scope: USA, Norway, Russia). "any" needs no VPN.
func vpnSupported(country string) bool {
	switch country {
	case "US", "NO", "RU":
		return true
	default:
		return false
	}
}

// StubGeoGate is a deterministic reference implementation for contract tests.
// It performs NO network I/O; the reachability results are configured.
type StubGeoGate struct {
	NativeCountry   string // the device's detected country
	ReachableNative bool   // reachable without VPN
	ReachableViaVPN bool   // reachable after connecting the VPN
}

func (g StubGeoGate) DetectCountry(context.Context) (string, error) {
	return g.NativeCountry, nil
}

func (g StubGeoGate) Reachable(context.Context, string) (bool, error) {
	return g.ReachableNative, nil
}

func (g StubGeoGate) Ensure(ctx context.Context, endpoint, requiredCountry string) (Decision, error) {
	country, err := g.DetectCountry(ctx)
	if err != nil {
		return Decision{}, err
	}
	d := Decision{Country: country}

	// Reachability FIRST — never burn a VPN hop when native works.
	if g.ReachableNative {
		d.Reachable = true
		d.Outcome = ProceedNative
		return d, nil
	}

	// Not reachable natively. VPN only when the required country is in scope.
	if vpnSupported(requiredCountry) && g.ReachableViaVPN {
		d.Reachable = true
		d.UsedVPN = true
		d.VPNCountry = requiredCountry
		d.Outcome = ProceedVPN
		return d, nil
	}

	// Genuinely blocked → SKIP-with-reason, never FAIL-for-geo.
	d.Outcome = SkipRestricted
	if vpnSupported(requiredCountry) {
		d.Reason = ReasonGeoRestricted // VPN attempted but still unreachable
	} else {
		d.Reason = ReasonUnreachableExt // no VPN path for this country
	}
	return d, nil
}
