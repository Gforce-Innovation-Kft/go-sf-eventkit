package eventkit_test

import (
	"testing"

	"github.com/Gforce-Innovation-Kft/go-sf-eventkit"
)

func TestAttributesCarryRoutingAndVersion(t *testing.T) {
	attrs := eventkit.Attributes(sample())
	for k, want := range map[string]string{
		eventkit.AttrVersion:  eventkit.Version,
		eventkit.AttrOrg:      "gabor-devhub",
		eventkit.AttrStream:   "LoginEventStream",
		eventkit.AttrCategory: "authentication",
	} {
		if got := attrs[k]; got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

// The map is handed to a propagator, which writes into it. If Attributes ever
// returned a shared or read-only map, trace context would be silently dropped
// and every trace would stop at the queue.
func TestAttributesIsWritableAndFresh(t *testing.T) {
	e := sample()
	first := eventkit.Attributes(e)
	first[eventkit.AttrTraceParent] = "00-4bf92f-00f067aa0ba902b7-01"

	second := eventkit.Attributes(e)
	if _, leaked := second[eventkit.AttrTraceParent]; leaked {
		t.Error("a second call saw the first call's trace context; the map is shared")
	}
}

func TestSupportedVersion(t *testing.T) {
	cases := map[string]struct {
		attrs map[string]string
		want  bool
	}{
		"current":             {map[string]string{eventkit.AttrVersion: eventkit.Version}, true},
		"absent":              {map[string]string{}, true},
		"from a future build": {map[string]string{eventkit.AttrVersion: "99"}, false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := eventkit.SupportedVersion(c.attrs); got != c.want {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}
