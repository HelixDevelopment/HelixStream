package registry

import (
	"errors"
	"testing"

	"digital.vasic.helixstream/pkg/adapter"
)

func TestRegisterAndLookupAdapter(t *testing.T) {
	r := New()
	a := adapter.NewStub("svc-a", adapter.Video, "any", nil, nil, nil)

	if err := r.RegisterAdapter(a); err != nil {
		t.Fatalf("RegisterAdapter: %v", err)
	}
	got, ok := r.Adapter("svc-a")
	if !ok || got.Name() != "svc-a" {
		t.Fatalf("Adapter(svc-a) = %v ok=%v", got, ok)
	}
	if _, ok := r.Adapter("nope"); ok {
		t.Fatal("Adapter(nope) should be absent")
	}
}

func TestRegisterAdapterErrors(t *testing.T) {
	r := New()
	if err := r.RegisterAdapter(nil); err == nil {
		t.Fatal("nil adapter must error")
	}
	if err := r.RegisterAdapter(adapter.NewStub("", adapter.Video, "any", nil, nil, nil)); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("empty name = %v, want ErrEmptyName", err)
	}
	a := adapter.NewStub("dup", adapter.Video, "any", nil, nil, nil)
	_ = r.RegisterAdapter(a)
	if err := r.RegisterAdapter(a); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate = %v, want ErrDuplicate", err)
	}
}

func TestNamesSorted(t *testing.T) {
	r := New()
	for _, n := range []string{"c", "a", "b"} {
		_ = r.RegisterAdapter(adapter.NewStub(n, adapter.Video, "any", nil, nil, nil))
	}
	got := r.Names()
	want := []string{"a", "b", "c"}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
}

func TestRoster(t *testing.T) {
	r := New()
	if err := r.RegisterRoster([]RosterEntry{{Package: "p.one", AppName: "One"}}); err != nil {
		t.Fatalf("RegisterRoster: %v", err)
	}
	if err := r.RegisterRoster([]RosterEntry{{Package: ""}}); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("empty-package roster = %v, want ErrEmptyName", err)
	}
	if got := r.Roster(); len(got) != 1 || got[0].Package != "p.one" {
		t.Fatalf("Roster() = %v", got)
	}
}

func TestEndpoints(t *testing.T) {
	r := New()
	if err := r.RegisterEndpoint("sink", "http://127.0.0.1:9000"); err != nil {
		t.Fatalf("RegisterEndpoint: %v", err)
	}
	if u, ok := r.Endpoint("sink"); !ok || u != "http://127.0.0.1:9000" {
		t.Fatalf("Endpoint(sink) = %q ok=%v", u, ok)
	}
	if err := r.RegisterEndpoint("", "x"); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("empty endpoint name = %v, want ErrEmptyName", err)
	}
	if err := r.RegisterEndpoint("sink", "y"); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate endpoint = %v, want ErrDuplicate", err)
	}
}
