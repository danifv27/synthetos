# synthetos test scenario

This command loads the specified Cucumber feature configuration, runs the tests, updates the Prometheus metric with the test results, and starts an HTTP server that listens on the specified port number.

### Options inherited from parent commands

| Name                       | Environment Variable | Default Value | Description |
| :--------------------------| :--------------------| :-------------| :-----------|
| --help (-h)                | Display help for the specified command. |
| --logging.level \<string>  | SC\_LOGGING\_LEVEL | info | Set the logging level (debug|info|warn|error|fatal) | 
| --logging.format \<string> | SC\_LOGGING\_OUTPUT_JSON | false | If set the log output is formatted as a JSON |
| --test.port \<integer> | SC\_TEST\_PORT | 8081 | The port number on which the HTTP server will listen for health probes. |
| --test.liveness \<boolean> | SC\_TEST\_LIVENESS | false | Enable/disable the liveness endpoint. |
| --test.readiness \<boolean> | SC\_TEST\_READINESS | false | Enable/disable the readiness endpoint. |

### Config file

```json
{
    "synthetos": {
        "test": {
            "scenario": {
            }
        }
    }
}
```