# atlas test scenario

This command loads the specified Cucumber feature configuration, runs the tests, updates the Prometheus metric with the test results, and starts an HTTP server that listens on the specified port number. 
This command implements the multi-target exporter pattern, so we advice to read the guide [Understanding and using the multi-target exporter pattern](https://prometheus.io/docs/guides/multi-target-exporter/) to get the general idea about the configuration.

There are two ways of querying the exporter:

    * Querying the exporter itself. It has its own metrics, available at `/metrics`. Those are metrics in the Prometheus format. They come from the exporter’s instrumentation and tell us about the state of the exporter itself while it is running.
    * Querying the exporter to test a scenario available at `/probe`. For this type of querying we need to provide feature name as parameter in the HTTP GET request. 

 &#x24D8;
 > Currently there is a limitation when using features. Each feature can have only one scenario defined

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
    "atlas": {
        "test": {
            "scenario": {
            }
        }
    }
}
```