# ECP Metrics Server

This application gathers information from an ECP network and exposes it in a Prometheus-compatible format. The use case for ECP Metrics Server is gathering more insight into an ECP network than what ECP doesn't expose out-of-the-box (e.g. the versions of registered ECP components).

## Exposed Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| ecp_component_version | gauge | The version of all registered ECP components | `code`, `org`, `type`, `version` |

## Authentication

The application uses a simple authentication mechanism by expecting a bearer token to be passed in the `Authentication` HTTP header. See below for how to configure the token to be passed.

## Configuration

The application is configured through the following environment variables:

| Variable | Description |
| -------- | ----------- |
| CD_DB_HOST | Hostname of the ECP Component Directory database |
| CD_DB_NAME | Name of the ECP Component Directory database |
| CD_DB_USER | Username to use for connecting to the Component Directory database |
| CD_DB_PASS | Password to use for connecting to the Component Directory database |
| AUTH_TOKEN | The bearer token consumers must pass in the HTTP `Authorization` header for authenticating to the ECP Metrics Server |
