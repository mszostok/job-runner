# Certs

This directory contains x509 certificates and associated private keys used by [cli](../cmd/cli) and [Agent](../cmd/agent).

The content of this directory shouldn't be published. The [`.gitignore`](.gitignore) files ensures that content generate by `make gen-certs` is ignored.

## Overview

The `make gen-certs` generates the following files:

| File name               | Description                                             |
|-------------------------|---------------------------------------------------------|
| `agent_ca_cert.pem`     | Agent's CA certificate. Used by Client to auth Agent.   |
| `agent_ca_key.pem`      | Agent's CA private key. **Store it securely.**          |
| `agent_cert.pem`        | Agent's certificate sent to the Client.                 |
| `agent_key.pem`         | Agent's private key. **Store it securely.**             |
| `client_ca_cert.pem`    | Client's CA cerficicate. Used by Agent to auth Clients. |
| `client_ca_key.pem`     | Client's CA private key. **Store it securely.**         |
| `Groot_client_cert.pem` | Groot's cerficicate. It has `user` role.                |
| `Groot_client_key.pem`  | Groot's private key. **Store it securely.**             |
| `Ricky_client_cert.pem` | Ricky's certificate. It has `admin` role.               |
| `Ricky_client_key.pem`  | Ricky's private key. **Store it securely.**             |
