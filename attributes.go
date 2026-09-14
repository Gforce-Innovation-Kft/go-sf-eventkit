package eventkit

// Pub/Sub message attributes. They exist so a consumer can filter and route
// without decoding the body, and so a trace can survive the queue.
//
// Attributes are cheap but not free: Pub/Sub allows 100 per message, keys up
// to 256 bytes and values up to 1024. Everything here is small and bounded by
// the stream catalog, except the trace keys, which are bounded by the W3C
// spec.
const (
	AttrVersion  = "version"  // Version, so an old consumer can refuse a new format
	AttrOrg      = "org"      // the org name from the collector's config
	AttrStream   = "stream"   // e.g. LoginEventStream
	AttrCategory = "category" // e.g. authentication

	// W3C Trace Context. The names are fixed by the specification and are
	// what every OpenTelemetry propagator reads and writes, so a publisher
	// and a consumer interoperate without agreeing on anything beyond this.
	//
	// This package does not inject or extract them: doing so would mean
	// importing OpenTelemetry, and a contract package's dependencies become
	// everyone's dependencies. Each side uses its own propagator against the
	// same map — see the README.
	AttrTraceParent = "traceparent"
	AttrTraceState  = "tracestate"
)

// Attributes returns the routing attributes for e.
//
// The returned map is fresh and meant to be written to: inject trace context
// into it before publishing.
func Attributes(e Event) map[string]string {
	return map[string]string{
		AttrVersion:  Version,
		AttrOrg:      e.Org,
		AttrStream:   e.Stream,
		AttrCategory: e.Category,
	}
}

// SupportedVersion reports whether a message carrying attrs is one this build
// understands. A message with no version attribute predates versioning and is
// treated as current, because rejecting it would lose data for no benefit.
func SupportedVersion(attrs map[string]string) bool {
	v, ok := attrs[AttrVersion]
	return !ok || v == Version
}
