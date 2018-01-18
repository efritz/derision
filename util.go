package main

import (
	"io"
	"io/ioutil"
)

func readAll(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return data, nil
}
