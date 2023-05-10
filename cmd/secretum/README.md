# Secretum. An Application To Interact With Different Key Manager Systems

## Kms List Fortanix Secrets

List Fortanix available secure objects. To use this command, you will need to provide your Fortanix API access key using the --kms.list.fortanix.api-key flag or by setting the corresponding environment variable SC_KMS_LIST_FORTANIX_API_KEY. 

You can optionally provide the group ID to be scanned as an argument. 

The --kms.list.output flag can be used to specify the output format. Logging can be controlled using the --logging.level and --logging.json flags, and actuator features can be controlled using the --probes.enable, --probes.address, and --probes.root-prefix flags.

Use the -h or --help flag to show context-sensitive help.

### Usage

```sh
secretum kms list fortanix secrets [--kms.list.fortanix.api-key=STRING] [<group-id>] [--kms.list.output=OUTPUT_FORMAT]
````

#### Arguments

    <group-id> (optional): Group ID to be scanned

### Flags

The following flags can be used to configure this command:

| **Flag**                    | **Description** | **Default Value** |
|-----------------------------|-----------------|-------------------|
| -h, --help                  |Show context-sensitive help. | |
| --kms.list.fortanix.api-key | Your Fortanix API access key. You can obtain this key by logging into your Fortanix account and navigating to the 'API Keys' page in the 'Settings' section. | |
| --kms.list.output           | The output format to use. Supported values are 'table', 'json', and 'text'. | table |
| --logging.level             | Set the logging level (debug, info, warn, error, fatal). | info |
| --logging.json              | If set, the log output is formatted as a JSON. | false |
| --probes.enable             | Enable the actuator. | false |
| --probes.address            | The address and port number of the actuator. | :8081 |
| --probes.root-prefix        | The prefix for the internal routes of web endpoints. | /actuator |

### Environment Variables

The following environment variables can be used to overwrite the values of certain flags:

| **Flag**                    | **Environment Variable**                             |
|-----------------------------|------------------------------------------------------|
| --kms.list.fortanix.api-key | SC_KMS_LIST_FORTANIX_API_KEY                         |
| --kms.list.output           | SC_KMS_LIST_OUTPUT                                   |
| --logging.level             | SC_LOGGING_LEVEL                                     |
| --logging.json              | SC_LOGGING_OUTPUT_JSON                               |
| --probes.enable             | SC_PROBES_ENABLE                                     |
| --probes.address            | SC_PROBES_ADDRESS                                    |
| --probes.root-prefix        | SC_PROBES_ROOT_PREFIX                                |
| --probes.root-prefix        | The prefix for the internal routes of web endpoints. |
