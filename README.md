# Gonvey

Gonvey is a simple reverse proxy. It has a very basic load balancing that consists of randomly forwarding requests to one of the endpoints that matches the requests' path, and is configurable.

<p align="center">
    <img src="images/logo.png" width="350"/>
</p>
<p align="center">
    <a href="#license">
        <img src="https://img.shields.io/badge/license-Apache-blue.svg?style=flat" />
    </a>
    <a href="https://goreportcard.com/report/github.com/Ullaakut/gonvey">
        <img src="https://goreportcard.com/badge/github.com/Ullaakut/gonvey" />
    </a>
    <a href="https://github.com/Ullaakut/gonvey/releases/latest">
        <img src="https://img.shields.io/github/release/Ullaakut/gonvey.svg?style=flat" />
    </a>
    <a href="https://travis-ci.org/Ullaakut/gonvey">
        <img src="https://travis-ci.org/Ullaakut/gonvey.svg?branch=master" />
    </a>
</p>

## How to run it

* `docker-compose up`

## Configuration

Gonvey is configured using the environment. The simplest way is to edit the environment variables in the `docker-compose.yml` file at the root of the repository.

### GONVEY_LOG_LEVEL

Sets the log level. Default value is `DEBUG`.

Examples: `DEBUG`, `INFO`, `WARNING`, `ERROR`, `FATAL`.

### `GONVEY_SERVER_PORT`

Sets the port used by the proxy. Default value is `8888`.

Can be any value between `1` and `65535`.

### `GONVEY_PROXY_MAP`

Sets the paths and endpoints that are bound within the proxy. For now, it's stored in a JSON-encoded string. Default value is `{"/bloggo":["http://app1:4242"],"/test":["http://app2:4243","http://app3:4244","http://app4:4245"]}`.

Note that paths are matched in a random order, so if a proxy map is like such for example:

`{"/test/deep/bind":["http://app2:4243"],"/test":["http://app1:4242"]}`

And a request comes in for `/test/deep/bind`, it might go to either `app1` or `app2`. (This is because maps are unordered in go)

Examples:

* `{"/test1":["http://app1:4242"]}`
* `{"/api/v1":["http://app1:4242"],"/api/v2":["http://app2:4243"],"/api/v3":["http://app3:4244"],"/api/v4":["http://app4:4245"]}`

## Screenshots

<p align="center">
    <img width="100%" src="images/logs.png">
</p>

## License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.

See the License for the specific language governing permissions and limitations under the License.
