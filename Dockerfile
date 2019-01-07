FROM golang:1.11 AS goapp-base

WORKDIR /go/src/github.com/jordanp/goapp

RUN \
    go get \
        github.com/golang/dep/cmd/dep \
        golang.org/x/lint/golint \
        github.com/jstemmer/go-junit-report \
        honnef.co/go/tools/... \
        github.com/client9/misspell/cmd/misspell \
        github.com/canthefason/go-watcher/cmd/watcher

COPY .ssh/* /root/.ssh/

FROM goapp-base as goapp-build

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only

COPY . .

ARG VERSION

RUN \
    mkdir -p build && \
    go build -ldflags "github.com/jordanp/goapp/cmd/goapp=-X=github.com/jordanp/goapp/app.VERSION=$VERSION" -o build/goapp github.com/jordanp/goapp/cmd/goapp/.

FROM debian:stretch-slim

RUN \
    apt-get update && \
    apt-get install -y ca-certificates && \
    rm -r /var/lib/apt/lists/*

ENTRYPOINT ["goapp"]

COPY --from=goapp-build /go/src/github.com/jordanp/goapp/build/* /usr/bin/

