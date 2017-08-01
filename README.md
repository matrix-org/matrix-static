Riot Static
===========

### Installation
`git clone` this repository and follow instructions at [getgb.io](https://getgb.io).
Use gb to build the project using `gb build` then execute it from the `bin/` directory.



### Usage
First you must create a config, there is a sample json file provided or you can use the helper binary `register-guest` to register a guest on a given homeserver and write an appropriate config file.

`register-guest` takes the following options:

`--config-file=` to specify the config file, defaulting to `./config.json`.

`--homeserver-url=` to specify the Homeserver URL to use, defaulting to `https://matrix.org`.



The main binary, `riot-static` exhibits the following controls:
Accepts `PORT=` env variable to determine what port to use, defaulting to port 8000 if one is not specified. Will panic if port is in use.

Accepts 3 command line arguments:

`--config-file=` to specify the config file, defaulting to `./config.json`.

`--enable-pprof` if set, enables the `/debug/pprof` endpoints for debugging.

`--enable-prometheus-metrics` if set, enables the `/metrics` endpoint for metrics.

`--num-workers=` to specify the number of worker goroutines to start, defaults to 32

`--public-serve-prefix=` to specify the router prefix to use for the user-facing html-serving routes, defaults to `/`



### Support

Currently hosted at https://stormy-bastion-98790.herokuapp.com/

Discussion Matrix Room is [#riot-static:matrix.org](https://matrix.to/#/#riot-static:matrix.org)