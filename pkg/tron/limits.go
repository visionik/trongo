package tron

// Internal safety limits to reduce worst-case CPU/memory usage on adversarial inputs.
//
// These are intentionally conservative defaults. If you need to process larger
// payloads, consider adding an exported Decoder API with configurable limits.
//
// NOTE: these are vars (not const) so tests can temporarily override them.
var (
	maxInputBytes = 10 << 20  // 10 MiB
	maxTokens     = 1_000_000 // hard cap on token count
	maxParseDepth = 1_000     // nested arrays/objects/class instantiations
	maxWalkDepth  = 1_000     // reflect graph depth for Marshal
)
