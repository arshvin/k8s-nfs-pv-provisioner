FROM golang:1.13.1-alpine AS builder

ENV CGO_ENABLED=0

COPY ./src/fake-provisioner.go /go/src/fake-provisioner/
COPY ./go.mod /go/src/fake-provisioner/
COPY ./go.sum /go/src/fake-provisioner/

WORKDIR /go/src/fake-provisioner/
RUN go install .

FROM alpine:3.10.2
COPY --from=builder /go/bin/fake-provisioner /app/
ENTRYPOINT ["/app/fake-provisioner"]
