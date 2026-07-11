// Package playback defines the reusable playback-manipulation contract
// (DESIGN.md §7): enumerate tracks, choose the HIGHEST video/audio track, seek
// to a (random) position, transport (pause/resume/stop), select a specific
// track. Each real capability carries a captured-evidence assertion in later
// PWUs; this package defines the contract + an in-memory stub.
//
// P1 NOTE (skeleton): the stub manipulates in-memory track state only. It drives
// NO player. The real driver + captured-evidence (§11.4.107 liveness,
// §8.1 smooth-AV) land in P7/P3c.
package playback

import (
	"context"
	"errors"
)

// TrackKind distinguishes the three selectable track families.
type TrackKind string

const (
	Video    TrackKind = "video"
	Audio    TrackKind = "audio"
	Subtitle TrackKind = "subtitle"
)

// Track is one selectable track. Rank orders quality (higher = better): for
// video it is resolution/bitrate rank, for audio channel/codec rank.
type Track struct {
	Kind  TrackKind
	ID    string
	Label string
	Rank  int
}

// Transport is a playback transport command.
type Transport string

const (
	Play   Transport = "play"
	Pause  Transport = "pause"
	Resume Transport = "resume"
	Stop   Transport = "stop"
)

var (
	// ErrNoTracks is returned by SelectHighest when there are no tracks of the kind.
	ErrNoTracks = errors.New("playback: no tracks of requested kind")
	// ErrSeekRange is returned by Seek when the fraction is outside [0,1].
	ErrSeekRange = errors.New("playback: seek fraction out of [0,1]")
	// ErrUnknownTrack is returned by SelectTrack for a track the controller does not offer.
	ErrUnknownTrack = errors.New("playback: unknown track")
)

// PlaybackController is the core playback-manipulation contract (one of the
// three ServiceAdapter sub-contracts).
type PlaybackController interface {
	// Tracks lists the available tracks of a kind.
	Tracks(ctx context.Context, kind TrackKind) ([]Track, error)
	// SelectHighest selects the max-Rank track of the kind. ErrNoTracks if none.
	SelectHighest(ctx context.Context, kind TrackKind) (Track, error)
	// Seek jumps to fraction ∈ [0,1] of the timeline. ErrSeekRange otherwise.
	Seek(ctx context.Context, fraction float64) error
	// Do issues a transport command.
	Do(ctx context.Context, t Transport) error
	// SelectTrack selects a specific track. ErrUnknownTrack if not offered.
	SelectTrack(ctx context.Context, t Track) error
}

// StubPlayback is an in-memory reference implementation for contract tests.
type StubPlayback struct {
	tracks   []Track
	Selected map[TrackKind]Track // last-selected track per kind (observable state)
	Position float64             // last seek fraction
	LastCmd  Transport
}

// NewStub builds a StubPlayback with the given track set.
func NewStub(tracks []Track) *StubPlayback {
	return &StubPlayback{tracks: tracks, Selected: map[TrackKind]Track{}}
}

func (p *StubPlayback) Tracks(_ context.Context, kind TrackKind) ([]Track, error) {
	var out []Track
	for _, t := range p.tracks {
		if t.Kind == kind {
			out = append(out, t)
		}
	}
	return out, nil
}

func (p *StubPlayback) SelectHighest(ctx context.Context, kind TrackKind) (Track, error) {
	ts, _ := p.Tracks(ctx, kind)
	if len(ts) == 0 {
		return Track{}, ErrNoTracks
	}
	best := ts[0]
	for _, t := range ts[1:] {
		if t.Rank > best.Rank {
			best = t
		}
	}
	p.Selected[kind] = best
	return best, nil
}

func (p *StubPlayback) Seek(_ context.Context, fraction float64) error {
	if fraction < 0 || fraction > 1 {
		return ErrSeekRange
	}
	p.Position = fraction
	return nil
}

func (p *StubPlayback) Do(_ context.Context, t Transport) error {
	p.LastCmd = t
	return nil
}

func (p *StubPlayback) SelectTrack(_ context.Context, t Track) error {
	for _, have := range p.tracks {
		if have.ID == t.ID && have.Kind == t.Kind {
			p.Selected[t.Kind] = have
			return nil
		}
	}
	return ErrUnknownTrack
}
