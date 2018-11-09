FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/consumers/metrics/config.yaml .
COPY ./bin/metrics /bin/metrics
ENTRYPOINT ["/bin/metrics"]