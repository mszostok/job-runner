## Linux Process Runner

Linux Process Runner (LPR) provides functionality to run arbitrary Linux processes. It consists of Agent (server) and Client (CLI) that communicate over gRPC.

## Prerequisites

* [Go](https://golang.org/dl/) 1.17 or higher
* Make

Helper scripts may introduce additional dependencies. However, all helper scripts support the `INSTALL_DEPS` environment variable flag.
By default, this flag is set to `false`. This way, the scripts will try to use the tools installed on your local machine. This helps speed up the development process.
If you do not want to install any additional tools, or you want to ensure reproducible script
results, export `INSTALL_DEPS=true`. This way, the proper tool version will be automatically installed and used.

## Dependency management

This project uses `go modules` for dependency management. To install all required dependencies, use the following command:

```bash
go mod download
```

## Testing

### Unit tests

To run all unit tests, execute:

```bash
make test-unit
```

To generate the unit test coverage HTML report, execute:

```bash
make test-cover-html
```

> **NOTE:** The generated report opens automatically in your default browser.

### Lint tests

To check your code for errors, such as typos, wrong formatting, security issues, etc., execute:

```bash
make test-lint
```

To automatically fix detected lint issues, execute:

```bash
make fix-lint-issues
```
