# SPDX-License-Identifier: MIT
FROM golang:1.22 as builder
WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY app.go app.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o app app.go

FROM ubuntu:22.04
RUN apt install -y curl yq jq sqlite
EXPOSE 8080 9090
WORKDIR /
COPY --from=builder /workspace/app .
COPY index.tmpl index.tmpl
USER 65532:65532
ENTRYPOINT ["/app"]