FROM golang:1.11-alpine

COPY . /src
WORKDIR /src

RUN apk --update add git
RUN go get github.com/constabulary/gb/...
RUN go get github.com/valyala/quicktemplate/qtc
RUN qtc
RUN gb build

FROM alpine

# We need this otherwise we don't have a good list of CAs
RUN apk --update add ca-certificates && mkdir /opt/matrix-static

WORKDIR /opt/matrix-static/
COPY --from=0 /src/bin/* /bin/

CMD ["matrix-static"]
