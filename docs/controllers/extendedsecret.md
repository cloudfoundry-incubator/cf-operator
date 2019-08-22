# ExtendedSecret

- [ExtendedSecret](#extendedsecret)
  - [Description](#description)
  - [Types](#types)
  - [Features](#features)
    - [Generated](#generated)
    - [Policies](#policies)
    - [Auto-approving Certificates](#auto-approving-certificates)
  - [`ExtendedSecret` Examples](#extendedsecret-examples)

## Description

## Types

`ExtendedSecret` supports generating the following:

| Secret Type                     | spec.type     | certificate.signerType | certificate.isCA    |
| ------------------------------- | ------------- | ---------------------- | ------------------- |
| `passwords`                     | `password`    | not set                | not set             |
| `rsa keys`                      | `rsa`         | not set                | not set             |
| `ssh keys`                      | `ssh`         | not set                | not set             |
| `self-signed root certificates` | `certificate` | `local`                | `true`              |
| `self-signed certificates`      | `certificate` | `local`                | `false`             |
| `cluster-signed certificates`   | `certificate` | `cluster`              | `false`             |

> **Note:**
>
> You can find more details in the [BOSH docs](https://bosh.io/docs/variable-types).

## Features

### Generated

A pluggable implementation for generating certificates and passwords.

### Policies

The developer can specify policies for rotation (e.g. automatic or not) and how secrets are created (e.g. password complexity, certificate expiration date, etc.).

### Auto-approving Certificates

A certificate `ExtendedSecret` can be signed by the Kube API Server. The ExtendedSecret Controller is be responsible for generating certificate signing request and approving the request:

```yaml
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: generate-certificate
spec:
  request: ((encoded-cert-signing-request))
  usages:
  - digital signature
  - key encipherment
```

The CertificateSigningRequest controller watches for `CertificateSigningRequest` and approves ExtendedSecret-owned CSRs and persists the generated certificate.

## `ExtendedSecret` Examples

See https://github.com/cloudfoundry-incubator/cf-operator/tree/master/docs/examples/extended-secret
