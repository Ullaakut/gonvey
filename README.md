# Gonvey

Gonvey is a simple reverse proxy.

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
