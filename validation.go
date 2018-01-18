package main

import (
	"encoding/json"
	"io"

	"github.com/xeipuuv/gojsonschema"
)

type (
	jsonHandler struct {
		Expectation expectation `json:"request"`
		Template    template    `json:"response"`
	}

	validationError struct {
		errors []gojsonschema.ResultError
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
		return nil, nil, &validationError{result.Errors()}
	}

	payload := &jsonHandler{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, nil, err
	}

	return &payload.Expectation, &payload.Template, nil
}

func (e *validationError) Error() string {
	return "validation error"
}
