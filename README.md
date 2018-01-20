# Derision

[![GoDoc](https://godoc.org/github.com/efritz/derision?status.svg)](https://godoc.org/github.com/efritz/derision)
[![Build Status](https://secure.travis-ci.org/efritz/derision.png)](http://travis-ci.org/efritz/derision)
[![Code Coverage](http://codecov.io/github/efritz/derision/coverage.svg?branch=master)](http://codecov.io/github/efritz/derision?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/efritz/derision)](https://goreportcard.com/report/github.com/efritz/derision)

A logging mock HTTP API for integration testing. Under development.

## Example

To start the server via docker, run the following command.

```bash
docker run --rm -it -p 5000:5000 efritz/derision
```

Then, before making any API requests to the server, register a set of request
expectations and their corresponding response template. The following example
creates a JSON response for a path matching a user route. Each field of the
response must be a valid [Go template](https://golang.org/pkg/text/template/)
which allows pulling portions of data from the request.

```bash
curl -H 'Content-Type: application/json' -d '{
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
}' http://localhost:5000/_control/register
```

Multiple expectations can be registered and are evaluated in-order. A request
to the API that matches the expectation will receive a response based on the
associated template. If a request does not match any expectation, the API will
respond with an empty 404. All expectations can be cleared by performing a `POST`
request to the `_control/clear` endpoint.

The following curl command illustrates the above expectation.

```bash
$ curl -v http://localhost:5000/users/123
< HTTP/1.1 200 OK
< Content-Type: application/json
< X-Request-Id: a1d2c5b8-b3d7-4e49-bd26-1e2dfee8eef5
< Date: Sat, 20 Jan 2018 17:18:26 GMT
< Content-Length: 37
<
{"user_id": 50, "username": "foobar"}
```

After making non-control requests to the API, the sequence of requests can be
listed via the following endpoint.

```
bash
$ curl -v http://localhost:5000/_control/requests
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Sat, 20 Jan 2018 17:24:22 GMT
< Content-Length: 106
<
[{
    "method": "GET",
    "path": "\/users\/123",
    "headers": {
        "Accept": ["*\/*"],
        "User-Agent": ["curl\/7.54.0"]
    },
    "body": ""
}]
```

Requesting this with the `?clear=true` query parameter will truncate this list.

## Response Templates

The following items are accessible within the response templates.

| Name         | Description |
| ------------ | ----------- |
| Method       | Raw request method |
| Path         | Raw request path |
| Headers      | Raw request headers (`string` to `[]string` pairs) |
| Body         | Raw request body |
| MethodGroups | Groups captured from the pattern match on the request method |
| PathGroups   | Groups captured from the pattern match on the request path |
| HeaderGroups | Groups captured from the pattern match on a request header value (`string` to `[]string` pairs) |

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
