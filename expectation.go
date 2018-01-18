package main

type (
	expectation struct {
		Method  string              `json:"method"`
		Path    string              `json:"path"`
		Headers map[string][]string `json:"headers"`
	}

	matcher func(*request) bool
)

func (e *expectation) Matches(r *request) bool {
	return matchAll(r, e.matchMethod, e.matchURL, e.matchHeaders)
}

func (e *expectation) matchMethod(r *request) bool {
	if e.Method == "" {
		return true
	}

	return r.Method == e.Method
}

func (e *expectation) matchURL(r *request) bool {
	if e.Path == "" {
		return true
	}

	return r.URL == e.Path
}

func (e *expectation) matchHeaders(r *request) bool {
	for k, match := range e.Headers {
		if !matchSlices(match, r.Headers[k]) {
			return false
		}
	}

	return true
}

//
// Helpers

func matchAll(r *request, matchers ...matcher) bool {
	for _, m := range matchers {
		if !m(r) {
			return false
		}
	}
	return true
}

func matchSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
