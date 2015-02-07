# Go binding for Elliptics

[Elliptics](https://github.com/reverbrain/elliptics) is an amazing, distributed, fault tolerant, key-value storage.
Use [Elliptics](https://github.com/reverbrain/elliptics) from **Go** language.

### Documentation

[![GoDoc](https://godoc.org/github.com/noxiouz/elliptics-go/elliptics?status.png)](https://godoc.org/github.com/noxiouz/elliptics-go/elliptics)

### CI status

Branch  | Build status | Coverage
------------- | ------------- | -------
master  | [![Build Status](https://travis-ci.org/noxiouz/elliptics-go.png?branch=master)](https://travis-ci.org/noxiouz/elliptics-go?branch=master) | [![Coverage Status](https://coveralls.io/repos/noxiouz/elliptics-go/badge.svg?branch=master)](https://coveralls.io/r/noxiouz/elliptics-go?branch=master)
## Installation

You should install `elliptics-client-dev` to build this one.
It could be installed from a [repository](http://repo.reverbrain.com)
or build from a [source](https://github.com/reverbrain/elliptics).

```
go get github.com/noxiouz/elliptics-go/elliptics
```

Specify the following environment variables, if libraries and heders are located in a non-standard location:

 * `CGO_CFLAGS` - C flags
 * `CGO_CPPFLAGS` - both C++/C flags
 * `CGO_CXXFLAGS` - C++ flags
 * `CGO_LDFLAGS` - linker flags

