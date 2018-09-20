# Gonvey

Gonvey is a simple reverse proxy.

[![Build Status](https://travis-ci.org/Ullaakut/gonvey.svg?branch=master)](https://travis-ci.org/Ullaakut/gonvey)

## Dependencies

### Docker build

* `docker`

### Manual binary build

* `dep`

## How to run it

### Binary

* `dep ensure`
* `go run gonvey.go`

### Docker image

* `docker build -t . gonvey`
* `docker run gonvey`
