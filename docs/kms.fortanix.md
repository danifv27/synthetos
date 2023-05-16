# atlas kms fortanix

Command used to manage Fortanix Key Management Systems.

Flags are inherited to the `kms fortanix` subcommands.

## Flags

The following flags can be used to configure this command:

| Name                                      | Environment Variable             | Default Value                     | Description |
| :-----------------------------------------| :--------------------------------| :---------------------------------| :-----------|
| --kms.fortanix.api-endpoint-url \<string> | SC_KMS_FORTANIX_API_ENDPOINT_URL | https://kms-test.adidas-group.com | The URL for the Fortanix API endpoint. Make sure to include the trailing slash. |
| --kms.fortanix.api-key \<string>          | SC_KMS_FORTANIX_API_KEY          |                                   | Your Fortanix API access key. You can obtain this key by logging into your Fortanix account and navigating to the 'API Keys' page in the 'Settings' section. |

# Available Commands

* [kms fortanix decrypt](./kms.fortanix.decrypt.md)
* [kms fortanix list](./kms.fortanix.list.md)
