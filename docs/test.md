# atlas test

If the command runs inside a Kubernetes cluster there is the option to provides a liveness and readiness endpoint.

## Liveness Endpoint

If the liveness flag is set, the application provides a liveness endpoint that returns an HTTP 200 OK status code as long as the application is running. The path of the liveness endpoint is `/liveness`.

## Readiness Endpoint

If the readiness flag is set, the application provides a readiness endpoint that returns an HTTP 200 OK status code if the Cucumber feature configuration has been successfully loaded and parsed.  The path of the readiness endpoint is `/readiness`.

## Options

| Flag                 | Environment Variable      | Default Value | Description |
| :--------------------| :-------------------------| :------------ | :---------- |
| --actuator.enable\<boolean> | SC\_TEST\_ACTUATOR\_ENABLE | true | Enable actuator?. |
| --actuator.addresss \<string> | SC\_TEST\_ACTUATOR\_ADDRESS | ":8081" | Actuator address. |

### Options inherited from parent commands

| Name                       | Environment Variable | Default Value | Description |
| :--------------------------| :--------------------| :-------------| :-----------|
| --help (-h)                | Display help for the specified command. |
| --logging.level \<string>  | SC\_LOGGING\_LEVEL | info | Set the logging level (debug|info|warn|error|fatal) | 
| --logging.format \<string> | SC\_LOGGING\_OUTPUT_JSON | false | If set the log output is formatted as a JSON |

### Config file

```json
{
    "synthetos": {
        "test": {
            "actuator": {
                "enable": true,
                "liveness": true,
                "readiness": false,
            }
        }
    }
}
```

# Available Subcommands
* [scenario](./test.scenario.md)