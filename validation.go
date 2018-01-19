package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/xeipuuv/gojsonschema"
)

type (
	jsonHandler struct {
		Expectation jsonExpectation `json:"request"`
		Template    template        `json:"response"`
	}

	jsonExpectation struct {
		Method  string            `json:"method"`
		Path    string            `json:"path"`
		Headers map[string]string `json:"headers"`
	}

	validationError struct {
		errors []string
	}
)

const handlerSchema = `
{
	"type": "object",
	"properties": {
		"request": {
			"type": "object",
			"properties": {
				"method": {
					"type": "string",
					"enum": ["GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"]
				},
				"path": {
					"type": "string"
				},
				"headers": {
					"type": "object"
				}
			},
			"additionalProperties": false
		},
		"response": {
			"type": "object",
			"properties": {
				"status_code": {
					"type": "string",
					"pattern": "^\\d{3}$"
				},
				"headers": {
					"type": "object"
				},
				"body": {
					"type": "string"
				}
			},
			"additionalProperties": false
		}
	},
	"additionalProperties": false,
	"required": [
		"request",
		"response"
	]
}
`

func readAndValidate(rc io.ReadCloser) (*expectation, *template, error) {
	input, err := readAll(rc)
	if err != nil {
		return nil, nil, err
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(handlerSchema),
		gojsonschema.NewStringLoader(string(input)),
	)

	if err != nil {
		return nil, nil, err
	}

	if !result.Valid() {
		errors := []string{}
		for _, err := range result.Errors() {
			errors = append(errors, fmt.Sprintf("%s: %s", err.Field(), err.Description()))
		}

		return nil, nil, &validationError{errors}
	}

	payload := &jsonHandler{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, nil, err
	}

	expectation, err := convertExpectation(payload.Expectation)
	if err != nil {
		return nil, nil, err
	}

	return expectation, &payload.Template, nil
}

func convertExpectation(e jsonExpectation) (*expectation, error) {
	methodRegex, err := compile(e.Method)
	if err != nil {
		return nil, &validationError{[]string{"method: illegal regex"}}
	}

	pathRegex, err := compile(e.Path)
	if err != nil {
		return nil, &validationError{[]string{"path: illegal regex"}}
	}

	headerRegexMap := map[string]*regexp.Regexp{}
	for header, value := range e.Headers {
		regex, err := compile(value)
		if err != nil {
			return nil, &validationError{[]string{fmt.Sprintf("header %s: illegal regex", header)}}
		}

		if regex != nil {
			headerRegexMap[header] = regex
		}
	}

	return &expectation{
		method:  methodRegex,
		path:    pathRegex,
		headers: headerRegexMap,
	}, nil
}

func compile(val string) (*regexp.Regexp, error) {
	if val == "" {
		return nil, nil
	}

	return regexp.Compile(val)
}

func (e *validationError) Error() string {
	return "validation error"
}
