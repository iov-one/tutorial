# Tutorial

This is a simple app to show how to build an application
using the weave framework. This is based on the
[weave starter kit](https://github.com/iov-one/weave-starter-kit)
base. Feel free to copy that repo to start your own application,
and read the documentation at https://weave.readthedocs.io

This repo will be accompanied by a tutorial text explaining how you can
build such an app, and is designed that you follow along with your
own copy from weave-starter-kit. This is not designed as production-ready code.

## Goal

Let's build an exchange. Distributed exchanges are a hot topic and require a
fair bit of data modelling. Let us build a custom module to handle trading,
and then integrate it with all the useful modules that weave gives us
for free. At the end we have a useful custom application, our own DEX.

A further tutorial will explain how to use
[iov core](https://github.com/iov-one/iov-core) to build your own frontend
application with secure key management, to interact with this DEX.

## Running the demo app

**TODO**

`make all` will test and build the application.

## Working with go.mod

To keep the CI deterministic, `make build`, `make install`, and `make test` all use the `-mod=readonly` flag,
which doesn't update any dependencies. If you want to update some deps locally, change the versions
in `go.mod`, run `make mod` to "tidy up" the mod file, and then run tests with `make tf`,
to update `go.sum` appropriately.
