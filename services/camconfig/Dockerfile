FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/camconfig/config.yaml .
COPY ./services/camconfig/web/templates ./web/templates
COPY ./bin/camconfig /bin/camconfig
ENTRYPOINT ["/bin/camconfig"]
