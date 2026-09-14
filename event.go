// Package eventkit is the contract between the collector and everything
// downstream of the queue: the shape of an event on the wire, and the message
// attributes that route and trace it.
//
// It has no dependencies, and deliberately so. Both sides of a queue must
// agree on this package, so anything it imports they must both import too. It
// holds types and encoding; it does no I/O and makes no decisions.
package eventkit

import (
	"encoding/json"
	"fmt"
	"time"
)

// Version is the wire format's version, carried in AttrVersion so a consumer
// can reject a message it does not understand rather than misread it.
const Version = "1"

// Event is one collected, decoded, policy-filtered Event Monitoring event.
//
// It mirrors the collector's internal event type, with two differences that
// matter: every field is explicitly tagged, and Category and ReplayMode are
// plain strings. A consumer must be able to read this without linking the
// collector.
type Event struct {
	// Provenance.
	Org      string `json:"org"`
	OrgID    string `json:"org_id"`
	Stream   string `json:"stream"`
	Topic    string `json:"topic"`
	Category string `json:"category"`

	// Identity. ID is the Salesforce EventIdentifier where the record has
	// one, and the Pub/Sub producer id otherwise. It is the deduplication
	// key: delivery is at-least-once, so a consumer will see repeats.
	ID     string    `json:"event_id"`
	Time   time.Time `json:"event_time,omitzero"`
	UserID string    `json:"user_id,omitempty"`

	// Transport. ReplayID is the position in the Salesforce stream; it is
	// opaque and only meaningful to a subscriber of that same stream.
	ReplayID   []byte    `json:"replay_id,omitempty"`
	SchemaID   string    `json:"schema_id,omitempty"`
	ReceivedAt time.Time `json:"received_at"`
	ReplayMode string    `json:"replay_mode,omitempty"`

	// Fields is the decoded payload after the collector's field policy has
	// been applied. In identity mode it holds three keys; in full mode it
	// holds every field the schema carries, with redacted values replaced by
	// the string "[redacted]" and their keys kept.
	Fields map[string]any `json:"fields,omitempty"`
}

// Encode marshals e for the wire.
func Encode(e Event) ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("encode event %s: %w", e.ID, err)
	}
	return b, nil
}

// Decode parses one wire message.
func Decode(b []byte) (Event, error) {
	var e Event
	if err := json.Unmarshal(b, &e); err != nil {
		return Event{}, fmt.Errorf("decode event: %w", err)
	}
	return e, nil
}

// Valid reports whether e carries the fields a consumer needs to do anything
// useful: something to deduplicate on, and enough provenance to route it.
//
// It deliberately does not check Fields. An event whose payload is empty is
// still a real event — the identity field policy plus a schema with nothing
// else populated produces exactly that — and dropping it would be a silent
// loss of a security signal.
func (e Event) Valid() error {
	switch {
	case e.ID == "":
		return fmt.Errorf("event has no id")
	case e.Org == "":
		return fmt.Errorf("event %s has no org", e.ID)
	case e.Stream == "":
		return fmt.Errorf("event %s has no stream", e.ID)
	}
	return nil
}
