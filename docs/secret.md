# secret command

This command allows you to interact with different Key Manager Systems (KMS). It provides a simple and easy-to-use interface to access and manage your secrets securely. With it, you can easily perform tasks like retrieving a secret, listing security objects, and decrypting a secret.

To use this command, simply provide the necessary authentication credentials for the KMS you want to interact with. Currently, the application supports Fortanix as the KMS provider. Once authenticated, you can use the different commands to interact with the KMS. For example, to retrieve a secret from Fortanix, you can use the get command and provide the secret ID. Similarly, you can use the list command to list all the security objects from a Fortanix group.

## Options inherited from parent commands

| Name                           | Environment Variable       | Default Value | Description                                                              |
| :------------------------------| :--------------------------| :-------------| :------------------------------------------------------------------------|
| --help (-h)                    |                            |               | Display help for the specified command.                                  |
| --logging.level \<string>      | SC\_LOGGING\_LEVEL         | info          | Set the logging level (debug|info|warn|error|fatal)                      | 
| --logging.format \<string>     | SC\_LOGGING\_OUTPUT_JSON   | false         | If set the log output is formatted as a JSON                             |
| --[no]-probes.enable\<boolean> | SC\_TEST\_PROBES_\_ENABLE  | true          | Enable actuator?                                                         |
| --probes.addresss \<string>    | SC\_TEST\_PROBES_\_ADDRESS | ":8081"       | Actuator address on which the HTTP server will listen for health probes. |

## Config file

```json
{
    "synthetos": {
        "secret": {
        }
    }
}
```