# go-sf-eventkit

The contract between the collector and everything downstream of the queue:
the shape of an event on the wire, and the message attributes that route and
trace it.

Part of [sf-event-platform](https://github.com/Gforce-Innovation-Kft/sf-event-platform).

## Zero dependencies, on purpose

Both sides of a queue must agree on this package, so **anything it imports,
they must both import**. It holds types and encoding: no I/O, no clients, no
decisions. If it ever grows behaviour, the split was wrong.

That includes OpenTelemetry. Trace context travels in the message attributes,
but this package does not inject or extract it — it only fixes the key names,
which the W3C specification already decided.

## Publishing

```go
attrs := eventkit.Attributes(e)
otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(attrs))

body, err := eventkit.Encode(e)
// publish body with attrs
```

`Attributes` returns a fresh, writable map precisely so the propagator can
write into it.

## Consuming

```go
if !eventkit.SupportedVersion(msg.Attributes) {
    // a newer publisher; do not guess at the format
}
ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(msg.Attributes))

e, err := eventkit.Decode(msg.Data)
if err == nil {
    err = e.Valid()
}
```

Both sides use their own propagator against the same map. Neither needs to
know what the other is running.

## What travels

| | |
|---|---|
| Body | JSON — debuggable in the console, and a sink can store it as-is |
| `event_id` | the deduplication key |
| `replay_id` | base64, because it is arbitrary bytes and a string field would corrupt it |
| `fields` | the payload *after* the collector's field policy — redacted values are `"[redacted]"` with the key kept |
| Attributes | `version`, `org`, `stream`, `category`, plus `traceparent` / `tracestate` |

**Deduplicate on `event_id`.** Delivery is at-least-once from Pub/Sub, and the
collector itself can re-deliver after a reconnect or a deploy, so repeats are
normal rather than exceptional.

**`Valid()` does not check `fields`.** An event with an empty payload is still
a real event — identity field policy against a sparse schema produces exactly
that — and dropping it would silently lose a security signal.

## Versioning

`version` rides in the attributes so a consumer can refuse a format it does
not understand instead of misreading it. A message with no version attribute
predates versioning and is treated as current: rejecting it would lose data
for no benefit.
