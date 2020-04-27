FROM golang:1.14-alpine AS build-stage
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add git make openssl

WORKDIR /go/src/github.com/4sh/k8s-scheduling-webhook
COPY . .

RUN make app

# Final image.
FROM alpine:latest
RUN apk --no-cache add \
  ca-certificates
COPY --from=build-stage /go/src/github.com/4sh/k8s-scheduling-webhook/scheduling-webhook /usr/local/bin/scheduling-webhook
ENTRYPOINT ["/usr/local/bin/scheduling-webhook"]