# atlas kube

Command used to interact with kubernetes clusters.

Flags are inherited to the `kube` subcommands.

The --kms.output flag can be used to specify the output format and actuator features can be controlled using the --kube.probes.enable, --kube.probes.address, and --kube.probes.root-prefix flags.

Use the -h or --help flag to show context-sensitive help.

## Flags

The following flags can be used to configure this command:

| Name                                 | Environment Variable  | Default Value | Description                                                                 |
| :------------------------------------| :---------------------| :-------------| :---------------------------------------------------------------------------|
| --[no]-kube.probes.enable \<bool>    | SC_PROBES_ENABLE      | false         | Enable the actuator.                                                        |
| --kube.probes.address \<string>      | SC_PROBES_ADDRESS     | :8081         | The address and port number of the actuator.                                |
| --kube. probes.root-prefix \<string> | SC_PROBES_ROOT_PREFIX | /actuator     | The prefix for the internal routes of web endpoints.                        |
| --kube.output \<string>              | SC_KUBE_OUTPUT        | table         | The output format to use. Supported values are 'table', 'json', and 'text'. |
| --kube.namespace \<string>           | SC_KUBE_NAMESPACE     |               | Path to the kubeconfig file to use for requests or host url                 |
| --kube.path \<string>                | SC_KUBE_CONFIG_PATH   |               | Name of the kubeconfig context to use                                       |
| --kube.context \<string>             | SC_KUBE_CONTEXT       |               | Selector (label query) to filter on                                         |
| -l, --kube.selector \<string>        | SC_KUBE_SELECTOR      |               | Output format                                                               |

# Available Commands

* [kube images](./kube.images.md)
* [kube resources](./kube.resources.md)
