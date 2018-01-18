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

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		data = data[n:]
	}

	return nil
}
