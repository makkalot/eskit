FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/consumerstore/config.yaml .
COPY ./bin/consumerstore /bin/consumerstore
COPY ./bin/consumerstoregw /bin/consumerstoregw
ENTRYPOINT ["/bin/consumerstore"]
