package playback

import (
	"context"
	"testing"
)

func fixture() *StubPlayback {
	return NewStub([]Track{
		{Kind: Video, ID: "v480", Label: "480p", Rank: 1},
		{Kind: Video, ID: "v1080", Label: "1080p", Rank: 3},
		{Kind: Video, ID: "v720", Label: "720p", Rank: 2},
		{Kind: Audio, ID: "a_stereo", Label: "Stereo", Rank: 1},
		{Kind: Audio, ID: "a_51", Label: "5.1", Rank: 2},
	})
}

func TestSelectHighest(t *testing.T) {
	p := fixture()
	ctx := context.Background()

	v, err := p.SelectHighest(ctx, Video)
	if err != nil || v.ID != "v1080" {
		t.Fatalf("highest video = %+v err=%v, want v1080", v, err)
	}
	a, err := p.SelectHighest(ctx, Audio)
	if err != nil || a.ID != "a_51" {
		t.Fatalf("highest audio = %+v err=%v, want a_51", a, err)
	}
	if _, err := p.SelectHighest(ctx, Subtitle); err != ErrNoTracks {
		t.Fatalf("highest subtitle = %v, want ErrNoTracks", err)
	}
}

func TestSeekRange(t *testing.T) {
	p := fixture()
	ctx := context.Background()
	cases := []struct {
		f       float64
		wantErr bool
	}{
		{0, false}, {0.5, false}, {1, false}, {-0.01, true}, {1.5, true},
	}
	for _, c := range cases {
		err := p.Seek(ctx, c.f)
		if (err != nil) != c.wantErr {
			t.Fatalf("Seek(%v) err=%v wantErr=%v", c.f, err, c.wantErr)
		}
		if err == nil && p.Position != c.f {
			t.Fatalf("Seek(%v) did not record position (%v)", c.f, p.Position)
		}
	}
}

func TestTransportAndSelectTrack(t *testing.T) {
	p := fixture()
	ctx := context.Background()
	if err := p.Do(ctx, Pause); err != nil || p.LastCmd != Pause {
		t.Fatalf("Do(Pause) err=%v last=%v", err, p.LastCmd)
	}
	if err := p.SelectTrack(ctx, Track{Kind: Video, ID: "v720"}); err != nil {
		t.Fatalf("SelectTrack(v720) err=%v", err)
	}
	if got := p.Selected[Video].ID; got != "v720" {
		t.Fatalf("selected video = %q, want v720", got)
	}
	if err := p.SelectTrack(ctx, Track{Kind: Video, ID: "nope"}); err != ErrUnknownTrack {
		t.Fatalf("SelectTrack(nope) = %v, want ErrUnknownTrack", err)
	}
}
