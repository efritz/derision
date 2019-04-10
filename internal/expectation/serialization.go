package expectation

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type jsonExpectation struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func Unmarshal(payload []byte) (Expectation, error) {
	e := &jsonExpectation{}
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload (%s)", err.Error())
	}

	methodRegex, err := compile(e.Method)
	if err != nil {
		return nil, fmt.Errorf("illegal method regex")
	}

	pathRegex, err := compile(e.Path)
	if err != nil {
		return nil, fmt.Errorf("illegal path regex")
	}

	headerRegexMap := map[string]*regexp.Regexp{}
	for header, value := range e.Headers {
		regex, err := compile(value)
		if err != nil {
			return nil, fmt.Errorf("illegal header regex")
		}

		if regex != nil {
			headerRegexMap[header] = regex
		}
	}

	bodyRegex, err := compile(e.Body)
	if err != nil {
		return nil, fmt.Errorf("illegal body regex")
	}

	return &expectation{
		method:  methodRegex,
		path:    pathRegex,
		headers: headerRegexMap,
		body:    bodyRegex,
	}, nil
}

func compile(val string) (*regexp.Regexp, error) {
	if val == "" {
		return nil, nil
	}

	return regexp.Compile(val)
}
