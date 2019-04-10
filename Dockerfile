FROM golang:1.12 AS builder
WORKDIR /derision
RUN mkdir /empty_config
COPY cmd ./cmd
COPY internal ./internal
COPY go.mod go.sum ./

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o derision ./cmd/derision

FROM scratch
ENV CONFIG_DIR /config
COPY --from=builder /empty_config /config
COPY --from=builder /derision/derision .
COPY schemas /schemas
ENTRYPOINT ["/derision"]
