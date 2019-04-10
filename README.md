# Derision

[![GoDoc](https://godoc.org/github.com/efritz/derision?status.svg)](https://godoc.org/github.com/efritz/derision)
[![Build Status](https://secure.travis-ci.org/efritz/derision.png)](http://travis-ci.org/efritz/derision)
[![Maintainability](https://api.codeclimate.com/v1/badges/289a6ddd42c61a92adcf/maintainability)](https://codeclimate.com/github/efritz/derision/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/289a6ddd42c61a92adcf/test_coverage)](https://codeclimate.com/github/efritz/derision/test_coverage)

A logging mock HTTP API for integration testing.

## Overview

Derision is an HTTP API that logs all incoming requests and can be configured
on-the-fly to respond in a particular way to requests matching an expectation.

Derision is packaged as a docker image. To start the server via docker, run
the following command.

```bash
docker run --rm -it -p 5000:5000 efritz/derision
```

To register an expectation, POST to the `/register` endpoint the expected HTTP
request structure along with a response template. These fields are explained
with more detail in the next section. Each endpoint that controls the behavior
of the API must set the `X-Derision-Control` header.

The following example creates a JSON response for a path matching a user route.

```bash
curl -H 'X-Derision-Control: true' -X POST -d '{
    "request": {
        "path": "/users/(\\d+)"
    },
    "response": {
        "headers": {
            "X-Request-ID": ["a1d2c5b8-b3d7-4e49-bd26-1e2dfee8eef5"],
            "Content-Type": ["application/json"]
        },
        "body": "{\"user_id\": {{ index (index .PathGroups 1) 1}}, \"username\": \"foobar\"}"
    }
}' http://localhost:5000/register
```

Multiple expectations can be registered and are evaluated in-order. A request
to the API (without the X-Derision-Control header set) that matches the
expectation will receive a response based on the associated template. If a
request does not match any expectation, the API will respond with an empty 404.
All such requests are logged for later inspection.

The following curl command illustrates the above expectation.

```bash
$ curl -v http://localhost:5000/users/123
< HTTP/1.1 200 OK
< Content-Length: 37
< Content-Type: application/json
< X-Request-Id: a1d2c5b8-b3d7-4e49-bd26-1e2dfee8eef5
< Date: Wed, 10 Apr 2019 22:57:14 GMT
<
{"user_id": 50, "username": "foobar"}
```

To retrieve the list of non-control requests made to the API, GET the `/requests`
endpoint. This will return a chronologically ordered list of requests, including
its method, path, headers, body, form, and file contents.

```
$ curl -H 'X-Derision-Control: true' http://localhost:5000/requests | jq
[
  {
    "method": "GET",
    "path": "/users/123",
    "headers": {
      "Accept": [
        "*/*"
      ],
      "User-Agent": [
        "curl/7.54.0"
      ]
    },
    "body": "",
    "raw_body": "",
    "form": {},
    "files": {},
    "raw_files": {}
  }
]
```

Use a query string containing `?clear=true` to truncate the request log. By
default, the log has an unbounded capacity and will record all requests. You can
change this default behavior `REQUEST_LOG_CAPACITY` environment variable in the
Docker command. If the log is bounded, then older requests will be pushed out of
the log when new requests are made.

Requests made to the API can also be *streamed* as they are made by users via the
`/sse` endpoint. Multiple users can subscribe to the same event stream without
conflict. This endpoint serves one
[server-sent event](https://en.wikipedia.org/wiki/Server-sent_events) for each request.
Only requests that are made to the API after subscribing to events will be seen (but
they would still be available in the log given the log has capacity and has not been
cleared).

```
curl -H 'X-Derision-Control: true' http://localhost:5000/sse
data:{"method": "GET", "path": "/test1", ...}

data:{"method": "GET", "path": "/test2", ...}

data:{"method": "GET", "path": "/test3", ...}
```

Control requests are not logged in either the request log or the request stream.

Expectations may change over time in a testing scenario. Instead of having to
restart the API container, all registered expectations can be removed by POSTing
to the `/clear` endpoint, as follows.

```bash
curl -H 'X-Derision-Control: true' -X POST http://localhost:5000/clear
```

## Expectations

A expectation consists of the fields `method`, `path`, `headers`, and `body`.
Method, path, and body are regular expressions, and headers is a map from strings
to regular expressions. Capturing groups are supported.

A request matches an expectation if the method, path, headers, and body of the
expectation respectively match the method, path, headers, and body of the request.

A response template consists of the fields `status_code`, `headers`, and `body`.
Each field of the response template must be a valid
[Go template](https://golang.org/pkg/text/template/) which allows pulling portions
of data from the request.

The following variables are accessible within the response templates.

| Name         | Description |
| ------------ | ----------- |
| Method       | Raw request method |
| Path         | Raw request path |
| Headers      | Raw request headers (`string` to `[]string` pairs) |
| Body         | Raw request body |
| MethodGroups | Groups captured from the pattern match on the request method |
| PathGroups   | Groups captured from the pattern match on the request path |
| HeaderGroups | Groups captured from the pattern match on a request header value (`string` to `[]string` pairs) |
| BodyGroups   | Groups captured form the pattern match on the request body |

## Static Configuration

Expectations can be registered from a directory on API startup. The recommended
way to do this is to either volume mount to the /config directory in the Docker
container or to create a custom Docker image that uses efritz/derision as a base
and ADD/COPY the files to the same directtory. For an example of the latter, see
[efritz/derision-ok](https://github.com/efritz/derision-ok).

Files are loaded in alphabetical order non-recursively from the configuration
directory. Each file in the directory must be in YAML format -- notice that as
YAML is a superset of JSON, they can also be JSON. Each file should contain a
top-level list of items, each one being the same structure as a payload to the
`/register` endpoint.

All endpoints behave the same whether or not the API was configured from static
files on startup. This means that expectations may change as the API is used.

## License

Copyright (c) 2018 Eric Fritz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
