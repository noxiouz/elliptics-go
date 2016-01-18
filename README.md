# Go binding for Elliptics  [![Build Status](https://travis-ci.org/noxiouz/elliptics-go.png?branch=master)](https://travis-ci.org/noxiouz/elliptics-go?branch=master) [![codecov.io](https://codecov.io/github/noxiouz/elliptics-go/coverage.svg?branch=master)](https://codecov.io/github/noxiouz/elliptics-go?branch=master)

[Elliptics](https://github.com/reverbrain/elliptics) is an amazing, distributed, fault tolerant, key-value storage.
Use [Elliptics](https://github.com/reverbrain/elliptics) from **Go** language.

### Documentation

[![GoDoc](https://godoc.org/github.com/noxiouz/elliptics-go/elliptics?status.png)](https://godoc.org/github.com/noxiouz/elliptics-go/elliptics)

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
