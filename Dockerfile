FROM golang:1.11.4 as build-env
# All these steps will be cached
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

ENV PROJECT "hub.ahiho.com/ahiho/squirrel-srv"
ENV APP "squirrel"

ENV RELEASE "0.2.9"
ENV COMMIT $(git rev-parse --short HEAD)
ENV BUILD_TIME $(date -u '+%Y-%m-%d_%H:%M:%S')

# Build the binary
RUN bash -c "CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags \"-s -w -X ${PROJECT}/pkg/version.Release=${RELEASE} -X ${PROJECT}/pkg/version.Commit=${COMMIT} -X ${PROJECT}/pkg/version.BuildTime=${BUILD_TIME}\" -o ${APP}"

FROM alpine
COPY --from=build-env /app/squirrel /go/bin/squirrel

ENV GRPC_PORT 9090
ENV HTTP_PORT 8080

EXPOSE $GRPC_PORT
EXPOSE $HTTP_PORT

RUN apk --update upgrade && \
    apk add sqlite && \
    rm -rf /var/cache/apk/*
# See http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ADD certs/ca-certificates.crt /etc/ssl/certs/

ADD migrations /go/bin/migrations
ADD certs /go/bin/certs
ADD data.db /go/bin

WORKDIR /go/bin

CMD ["./squirrel"]

