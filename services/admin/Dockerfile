FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/admin/config.yaml .
COPY ./services/admin/templates ./templates
COPY ./services/admin/static ./static
COPY ./bin/admin /bin/admin
ENTRYPOINT ["/bin/admin"]
