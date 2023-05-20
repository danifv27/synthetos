# synthetos kms fortanix decrypt manifests

Parse a bunch of yaml manifests. If it found a `v1/Secret` with *fortanixGroupId* and *fortanixSecretName* annotations. It fetch the data from Fortanix KMS and replace the data section with the content fetched. The rest of the yaml objects are unmodified.

```sh
secretum kms fortanix decrypt manifests [--kms.fortanix.list.api-key=STRING] <input-path>
```
## Arguments

    <input-path>: Input file or - for stdin
