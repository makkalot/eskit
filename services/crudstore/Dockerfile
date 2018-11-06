FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/crudstore/config.yaml .
COPY ./bin/crudstore /bin/crudstore
COPY ./bin/crudstoregw /bin/crudstoregw
ENTRYPOINT ["/bin/crudstore"]