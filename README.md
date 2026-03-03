# Terraform Provider for Hetzner Robot

A [Terraform](https://www.terraform.io) provider for managing resources via
the [Hetzner Robot API](https://robot.your-server.de/doc/webservice/en.html).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Authentication

The provider uses HTTP basic authentication with the Hetzner Robot API. Credentials can be provided via provider
configuration or environment variables:

```shell
export HETZNER_ROBOT_USERNAME="your-robot-username"
export HETZNER_ROBOT_PASSWORD="your-robot-password"
```

## Building The Provider

```shell
go install
```

## Developing the Provider

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

To run the full suite of acceptance tests, run `make testacc`.

**Note:** Acceptance tests create real resources and require valid Hetzner Robot API credentials.

```shell
make testacc
```
