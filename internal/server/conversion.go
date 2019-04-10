package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/efritz/derision/internal/request"
)

func convertRequest(r *http.Request) (*request.Request, error) {
	defer r.Body.Close()
	buffer := &bytes.Buffer{}
	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, buffer))

	if err := tryParse(r); err != nil {
		return nil, fmt.Errorf("failed to parse form (%s)", err.Error())
	}

	files, rawFiles, err := parseFiles(r)
	if err != nil {
		return nil, err
	}

	snapshot := &request.Request{
		Method:   r.Method,
		Path:     r.URL.Path,
		Headers:  r.Header,
		Body:     buffer.String(),
		RawBody:  encode(buffer.String()),
		Form:     r.Form,
		Files:    files,
		RawFiles: rawFiles,
	}

	return snapshot, nil
}

func tryParse(r *http.Request) error {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct == "multipart/form-data" {
		return r.ParseMultipartForm(2e7)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	if _, err := ioutil.ReadAll(r.Body); err != nil {
		return err
	}

	return nil
}

func parseFiles(r *http.Request) (map[string]string, map[string]string, error) {
	files := map[string]string{}
	rawFiles := map[string]string{}

	if r.MultipartForm == nil {
		return files, rawFiles, nil
	}

	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to open file header (%s)", err.Error())
			}

			defer file.Close()

			content, err := ioutil.ReadAll(file)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read file (%s)", err.Error())
			}

			files[header.Filename] = string(content)
			rawFiles[header.Filename] = encode(string(content))
		}
	}

	return files, rawFiles, nil
}

func encode(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
