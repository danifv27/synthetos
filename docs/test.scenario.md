# synthetos test scenario

This command loads the specified Cucumber feature configuration, runs the tests, updates the Prometheus metric with the test results, and starts an HTTP server that listens on the specified port number.

### Options inherited from parent commands

| Name                       | Environment Variable | Default Value | Description |
| :--------------------------| :--------------------| :-------------| :-----------|
| --help (-h)                | Display help for the specified command. |
| --logging.level \<string>  | SC\_LOGGING\_LEVEL | info | Set the logging level (debug|info|warn|error|fatal) | 
| --logging.format \<string> | SC\_LOGGING\_OUTPUT_JSON | false | If set the log output is formatted as a JSON |
| --actuator.enable\<boolean> | SC\_TEST\_ACTUATOR\_ENABLE | true | Enable actuator?  |
| --actuator.addresss \<string> | SC\_TEST\_ACTUATOR\_ADDRESS | ":8081" | Actuator address on which the HTTP server will listen for health probes. |

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