FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/eventstore/config.yaml .
COPY ./bin/eventstore /bin/eventstore
COPY ./bin/eventstoregw /bin/eventstoregw
ENTRYPOINT ["/bin/eventstore"]