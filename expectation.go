package main

type (
	expectation struct {
		Method  string              `json:"method"`
		Path    string              `json:"path"`
		Headers map[string][]string `json:"headers"`
	}
)

func (e *expectation) Matches(r *request) bool {
	if e.Method != "" && r.Method != e.Method {
		return false
	}

	if e.Path != "" && r.URL != e.Path {
		return false
	}

	for k, match := range e.Headers {
		vals, ok := r.Headers[k]
		if !ok {
			return false
		}

		if len(vals) != len(match) {
			return false
		}

		for i, v := range vals {
			if v != match[i] {
				return false
			}
		}
	}

	return true
}
