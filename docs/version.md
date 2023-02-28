# synthetos version

The version command provides information about the current version of the software being used. It displays information such as the version number, build number, release date, and other relevant information. In addition, it may also provide information about the underlying operating system or platform that the software is running on.

Overall, the version command is a helpful tool for developers and users to keep track of the current version of the software and to ensure that they are using the latest and most up-to-date version of the application.

## Options
| Flag                 | Environment Variable      | Default Value | Description |
| :--------------------| :-------------------------| :------------ | :---------- |
| --version.output \<string> | SC\_VERSION\_OUTPUT | pretty | Specify the output format. Valid options are text (default) or json. (pretty\|json). |


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
        "version": {
            "logging": {
                "level": "debug",
                "format": "json"
            }
        }
    }
}
```