package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

func makeHandlersFromFile(path string) ([]handler, error) {
	if path == "" {
		return nil, nil
	}

	fmt.Printf("reading registration file %s\n", path)

	chunks, err := readFile(path)
	if err != nil {
		return nil, err
	}

	handlers := []handler{}
	for _, chunk := range chunks {
		handler, err := makeHandler(ioutil.NopCloser(bytes.NewReader(chunk)))
		if err != nil {
			return nil, err
		}

		handlers = append(handlers, handler)
	}

	return handlers, nil
}

// Attempt to read a file formatted as a JSON array. Will return
// a byte slice for each element of the array on success and the
// error on read or decode failure.
func readFile(path string) ([][]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	jsonContent, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize file (%s)", err.Error())
	}

	payload := []interface{}{}
	if err := json.Unmarshal(jsonContent, &payload); err != nil {
		return nil, fmt.Errorf("failed to deserialize file (%s)", err.Error())
	}

	chunks := [][]byte{}
	for _, item := range payload {
		chunk, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
