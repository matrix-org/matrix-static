run: build
	./basicserver

build: update
	go build -o basicserver

update:
	go get -u github.com/valyala/fasthttp
	go get -u github.com/valyala/quicktemplate/qtc

generate: update
	go generate

all: update generate build run
