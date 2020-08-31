FROM golang:1.13.1-alpine AS builder

ENV CGO_ENABLED=0

WORKDIR /build/
COPY ./go.* /build/
RUN go mod download

COPY . /build/

RUN go test -v ./...
RUN go install k8s-pv-provisioner/cmd/provisioner

FROM alpine:3.10.2
COPY --from=builder /go/bin/provisioner /app/
ENTRYPOINT ["/app/provisioner"]
