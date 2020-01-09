FROM golang:1.13.1-alpine AS builder

ENV CGO_ENABLED=0

COPY . /build/

WORKDIR /build/
RUN go test k8s-pv-provisioner/cmd/provisioner && go install k8s-pv-provisioner/cmd/provisioner

FROM alpine:3.10.2
COPY --from=builder /go/bin/provisioner /app/
ENTRYPOINT ["/app/provisioner"]
