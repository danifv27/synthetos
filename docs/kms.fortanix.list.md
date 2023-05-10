# atlas kms list fortanix

List different Fortanix objects.

Flags are inherited to the `kms list fortanix` subcommands.

## Flags

The following flags can be used to configure this command:

| Name                                           | Environment Variable                  | Default Value | Description |
| :----------------------------------------------| :-------------------------------------| :-------------| :-----------|
| --kms.fortanix.list.api-endpoint-url \<string> | SC_KMS_FORTANIX_LIST_API_ENDPOINT_URL | https://kms-test.adidas-group.com | The URL for the Fortanix API endpoint. Make sure to include the trailing slash. |
| --kms.fortanix.list.api-key \<string>          | SC_KMS_FORTANIX_LIST_API_KEY          | Your Fortanix API access key. You can obtain this key by logging into your Fortanix account and navigating to the 'API Keys' page in the 'Settings' section. |

# Available Commands

* [kms fortanix list groups](./kms.fortanix.list.groups.md)
* [kms fortanix list secrets](./kms.fortanix.list.secrets.md)