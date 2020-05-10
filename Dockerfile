FROM golang:1.13-alpine

RUN apk --update add git gcc musl-dev
RUN go get github.com/valyala/quicktemplate/qtc

WORKDIR /src
COPY . /src

RUN qtc
RUN go build ./cmd/matrix-static

FROM alpine

# We need this otherwise we don't have a good list of CAs
RUN apk --update add ca-certificates && mkdir /opt/matrix-static

WORKDIR /opt/matrix-static/
COPY --from=0 /src/matrix-static /bin/
COPY ./assets/ ./assets

ENTRYPOINT matrix-static
