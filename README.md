Matrix Static
===========

### Installation
`git clone` or download this repository as an archive and extract then follow below instructions.

```
go get github.com/constabulary/gb/...
go get github.com/valyala/quicktemplate/qtc
qtc
gb build
```
After this, executables will be in the `bin` directory.


### Usage
First you must create a config, there is a sample json file provided or you can use the helper binary `register-guest` to register a guest on a given homeserver and write an appropriate config file.

`register-guest` takes the following options:

`--config-file=` to specify the config file, defaulting to `./config.json`.

`--homeserver-url=` to specify the Homeserver URL to use, defaulting to `https://matrix.org`.



The main binary, `matrix-static` exhibits the following controls:

Accepts `PORT=` env variable to determine what port to use, defaulting to port 8000 if one is not specified. Will panic if port is in use.

Accepts the following command line arguments:

`--config-file=` to specify the config file, defaulting to `./config.json`.

`--enable-pprof` if set, enables the `/debug/pprof` endpoints for debugging.

`--enable-prometheus-metrics` if set, enables the `/metrics` endpoint for metrics.

`--num-workers=` to specify the number of worker goroutines to start, defaults to 32

`--public-serve-prefix=` to specify the router prefix to use for the user-facing html-serving routes, defaults to `/`



### Support

Currently hosted at https://view.matrix.org

Discussion Matrix Room is [#matrix-static:matrix.org](https://matrix.to/#/#matrix-static:matrix.org)
