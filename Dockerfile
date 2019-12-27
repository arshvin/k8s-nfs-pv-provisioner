FROM golang:1.13.1-alpine AS builder

ENV CGO_ENABLED=0

COPY ./src/provisioner.go /go/src/nfs-provisioner/
COPY ./go.mod /go/src/nfs-provisioner/
COPY ./go.sum /go/src/nfs-provisioner/

WORKDIR /go/src/nfs-provisioner/
RUN go install .

FROM alpine:3.10.2
COPY --from=builder /go/bin/nfs-provisioner /app/
ENTRYPOINT ["/app/nfs-provisioner"]
