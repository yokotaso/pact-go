package dsl

// Request is the default implementation of the Request interface.
// TODO: not sure I like the duality of the types below. Probably
// lift into specific types that capture it more clearly.
type Request struct {
	// Method is the HTTP method for the interaction (GET, PUT, POST etc.)
	Method string `json:"method"`

	// Path represents the URI part after the domain and port,
	// and may be a string or a simple matcher.
	Path interface{} `json:"path"`

	// Path represents the query string of the URI.
	Query string `json:"query,omitempty"`

	// Headers are represented as a map in key/value form, where value may be a
	// string or simple matcher.
	Headers map[string]interface{} `json:"headers,omitempty"`

	// Body is the response body for the interaction, provided as a string or
	// a PactBodyBuilder
	Body interface{} `json:"body,omitempty"`
}
