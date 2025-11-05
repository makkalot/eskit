FROM alpine:3.8

WORKDIR /go/src/github.com/makkalot/eskit/
COPY ./services/users/config.yaml .
COPY ./bin/users /bin/users
ENTRYPOINT ["/bin/users"]
