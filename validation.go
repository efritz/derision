package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/xeipuuv/gojsonschema"
)

type jsonHandler struct {
	Expectation expectation `json:"request"`
	Template    template    `json:"response"`
}

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
		for _, err := range result.Errors() {
			fmt.Printf("Validation error (%s: %s)\n", err.Field(), err.Description())
		}

		return nil, nil, fmt.Errorf("validation error")
	}

	payload := &jsonHandler{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, nil, err
	}

	return &payload.Expectation, &payload.Template, nil
}
