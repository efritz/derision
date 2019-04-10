package request

type Request struct {
	Method   string              `json:"method"`
	Path     string              `json:"path"`
	Headers  map[string][]string `json:"headers"`
	Body     string              `json:"body"`
	RawBody  string              `json:"raw_body"`
	Form     map[string][]string `json:"form"`
	Files    map[string]string   `json:"files"`
	RawFiles map[string]string   `json:"raw_files"`
}
