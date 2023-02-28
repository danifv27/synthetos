# synthetos test

If the command runs inside a Kubernetes cluster there is the option to provides a liveness and readiness endpoint.

## Liveness Endpoint

If the liveness flag is set, the application provides a liveness endpoint that returns an HTTP 200 OK status code as long as the application is running. The path of the liveness endpoint is `/liveness`.

## Readiness Endpoint

If the readiness flag is set, the application provides a readiness endpoint that returns an HTTP 200 OK status code if the Cucumber feature configuration has been successfully loaded and parsed.  The path of the readiness endpoint is `/readiness`.

## Options

| Flag                 | Environment Variable      | Default Value | Description |
| :--------------------| :-------------------------| :------------ | :---------- |
| --test.port \<integer> | SC\_TEST\_PORT | 8081 | The port number on which the HTTP server will listen for health probes. |
| --test.liveness \<boolean> | SC\_TEST\_LIVENESS | false | Enable/disable the liveness endpoint. |
| --test.readiness \<boolean> | SC\_TEST\_READINESS | false | Enable/disable the readiness endpoint. |

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
            "port": 8081,
            "liveness": true,
            "readiness": false,
        }
    }
}
```

# Available Subcommands
* [scenario](./test.scenario.md)