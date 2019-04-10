package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/efritz/derision/internal/expectation"
	"github.com/efritz/derision/internal/handler"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/derision/internal/template"
	"github.com/efritz/response"
	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"
)

type jsonHandler struct {
	Expectation json.RawMessage `json:"request"`
	Template    json.RawMessage `json:"response"`
}

var schemaPath = "/schemas"

func makeHandler(input []byte) (handler.Handler, error) {
	payload := &jsonHandler{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload (%s)", err.Error())
	}

	expectation, err := expectation.Unmarshal(payload.Expectation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal expectation (%s)", err.Error())
	}

	template, err := template.Unmarshal(payload.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal template (%s)", err.Error())
	}

	handler := func(r *request.Request) (response.Response, error) {
		if match := expectation.Matches(r); match != nil {
			return template.Respond(r, match)
		}

		return nil, nil
	}

	return handler, nil
}

func loadHandlers(handlerSet handler.HandlerSet, path string) error {
	schema, err := getSchema()
	if err != nil {
		return err
	}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read config directory")
	}

	for _, info := range infos {
		if !info.IsDir() {
			handlers, err := makeHandlersFromPath(schema, path, info.Name())
			if err != nil {
				return fmt.Errorf("failed to load handlers from %s (%s)", info.Name(), err.Error())
			}

			for _, handler := range handlers {
				handlerSet.Add(handler)
			}
		}
	}

	return nil
}

func getSchema() (*gojsonschema.Schema, error) {
	data, err := loadYAML(filepath.Join(schemaPath, "handlers.yaml"))
	if err != nil {
		return nil, err
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(data))
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func makeHandlersFromPath(schema *gojsonschema.Schema, segments ...string) ([]handler.Handler, error) {
	data, err := loadYAML(segments...)
	if err != nil {
		return nil, err
	}

	result, err := schema.Validate(gojsonschema.NewStringLoader(string(data)))
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		details := []string{}
		for _, err := range result.Errors() {
			details = append(details, fmt.Sprintf("%s: %s", err.Field(), err.Description()))
		}

		return nil, fmt.Errorf("invalid config: %s", strings.Join(details, ", "))
	}

	handlers, err := makeHandlers(data)
	if err != nil {
		return nil, err
	}

	return handlers, nil
}

func makeHandlers(input []byte) ([]handler.Handler, error) {
	payloads := []json.RawMessage{}
	if err := json.Unmarshal(input, &payloads); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload (%s)", err.Error())
	}

	handlers := []handler.Handler{}
	for _, payload := range payloads {
		handler, err := makeHandler(payload)
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, handler)
	}

	return handlers, nil
}

func loadYAML(segments ...string) ([]byte, error) {
	content, err := ioutil.ReadFile(filepath.Join(segments...))
	if err != nil {
		return nil, err
	}

	data, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, err
	}

	return data, nil
}
