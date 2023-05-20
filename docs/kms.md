# synthetos kms

Command used to manage different Key Management Systems. Right now only Fortanix KMS is supported.

Flags are inherited to the `kms` subcommands.

The --kms.output flag can be used to specify the output format and actuator features can be controlled using the --probes.enable, --probes.address, and --probes.root-prefix flags.

Use the -h or --help flag to show context-sensitive help.

## Flags

The following flags can be used to configure this command:

| Name                           | Environment Variable  | Default Value | Description |
| :------------------------------| :---------------------| :-------------| :-----------|
| --probes.enable \<bool>        | SC_PROBES_ENABLE      | false         | Enable the actuator. |
| --probes.address \<string>     | SC_PROBES_ADDRESS     | :8081         | The address and port number of the actuator. |
| --probes.root-prefix \<string> | SC_PROBES_ROOT_PREFIX | /actuator     | The prefix for the internal routes of web endpoints. |
| --kms.output \<string>         | SC_KMS_OUTPUT         | table         | The output format to use. Supported values are 'table', 'json', and 'text'. |

# Available Commands

* [kms fortanix](./kms.fortanix.md)
