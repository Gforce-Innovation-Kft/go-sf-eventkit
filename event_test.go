package eventkit_test

import (
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/Gforce-Innovation-Kft/go-sf-eventkit"
)

func sample() eventkit.Event {
	return eventkit.Event{
		Org: "example-org", OrgID: "00D000000000000EAA",
		Stream: "LoginEventStream", Topic: "/event/LoginEventStream",
		Category:   "authentication",
		ID:         "5a2c1e6b-bfa2-4f7c-83cb-4aad246c29ce",
		Time:       time.Date(2026, 9, 14, 10, 8, 20, 0, time.UTC),
		UserID:     "005000000000000AAA",
		ReplayID:   []byte{0, 0, 0, 0, 0, 0, 176, 68},
		SchemaID:   "Ow_p0ZSXHeyq3O8mqmKtFA",
		ReceivedAt: time.Date(2026, 9, 14, 10, 8, 23, 0, time.UTC),
		ReplayMode: "latest",
		Fields: map[string]any{
			"EventIdentifier": "5a2c1e6b-bfa2-4f7c-83cb-4aad246c29ce",
			"UserId":          "005000000000000AAA",
			"LoginKey":        "[redacted]",
		},
	}
}

func TestRoundTripPreservesEveryField(t *testing.T) {
	want := sample()
	b, err := eventkit.Encode(want)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := eventkit.Decode(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.ID != want.ID || got.Org != want.Org || got.Stream != want.Stream {
		t.Errorf("identity lost: got %+v", got)
	}
	if !got.Time.Equal(want.Time) || !got.ReceivedAt.Equal(want.ReceivedAt) {
		t.Errorf("timestamps lost: got %v %v", got.Time, got.ReceivedAt)
	}
	if string(got.ReplayID) != string(want.ReplayID) {
		t.Errorf("replay id lost: got %v want %v", got.ReplayID, want.ReplayID)
	}
	if !maps.Equal(stringify(got.Fields), stringify(want.Fields)) {
		t.Errorf("fields lost: got %v want %v", got.Fields, want.Fields)
	}
}

// A replay id is arbitrary bytes, not text. Base64 in JSON is what keeps it
// intact; a string field would corrupt it and the corruption would only show
// up later, as a rejected cursor.
func TestReplayIDSurvivesNonUTF8Bytes(t *testing.T) {
	e := sample()
	e.ReplayID = []byte{0xff, 0xfe, 0x00, 0x80, 0x01}
	b, err := eventkit.Encode(e)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := eventkit.Decode(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(got.ReplayID) != string(e.ReplayID) {
		t.Errorf("got %v want %v", got.ReplayID, e.ReplayID)
	}
}

func TestZeroEventTimeIsOmitted(t *testing.T) {
	e := sample()
	e.Time = time.Time{}
	b, err := eventkit.Encode(e)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if strings.Contains(string(b), "event_time") {
		t.Errorf("zero event_time should be omitted, got %s", b)
	}
}

func TestDecodeRejectsGarbage(t *testing.T) {
	if _, err := eventkit.Decode([]byte("{not json")); err == nil {
		t.Fatal("want an error for malformed JSON, got nil")
	}
}

func TestValid(t *testing.T) {
	for name, mutate := range map[string]func(*eventkit.Event){
		"no id":     func(e *eventkit.Event) { e.ID = "" },
		"no org":    func(e *eventkit.Event) { e.Org = "" },
		"no stream": func(e *eventkit.Event) { e.Stream = "" },
	} {
		t.Run(name, func(t *testing.T) {
			e := sample()
			mutate(&e)
			if err := e.Valid(); err == nil {
				t.Errorf("want an error for %q, got nil", name)
			}
		})
	}

	t.Run("empty fields is valid", func(t *testing.T) {
		e := sample()
		e.Fields = nil
		if err := e.Valid(); err != nil {
			t.Errorf("an event with no payload is still an event: %v", err)
		}
	})
}

func stringify(m map[string]any) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		s, _ := v.(string)
		out[k] = s
	}
	return out
}
