# Command-Line Interface

Comprehensive Swiss Army Knife for DevOps and SRE teams, featuring a variety of tools. Named after Atlas, the Greek titan who carried the weight of the world on his shoulders, this name suggests this app can handle any task. It supports several commands, each accessible through the application binary. It can be configured in three different ways:

* Using a configuration file in Json format. This file should be placed in one of these paths: `/etc/<binary_name>.json`, `$HOME/.<binary_name>.json` or `<BINARY PATH>/.<binary_name>.json`
* Through commands flags.
* Via environment variables

ℹ️ command flags have precedence over environment values

This cli has the particularity of building diferent binaries with diferent functionality. Right now we have available three different binaries:

* atlas: contains all the commands.
* uxperi: user experience exporter commands.
* secretum: kms management commands.

## Flags

ℹ️ Flags are inherited from parent commands.

CLI commands support both local (specific to the given command) and global (works for every command available) flags. Some of the most common global flags are:

| Name                       | Environment Variable | Default Value | Description |
| :--------------------------| :--------------------| :-------------| :-----------|
| --help (-h)                | Display help for the specified command. |
| --logging.level \<string>  | SC\_LOGGING\_LEVEL | info | Set the logging level (debug|info|warn|error|fatal) | 
| --logging.format \<string> | SC\_LOGGING\_OUTPUT_JSON | false | If set the log output is formatted as a JSON |

### Config file

```json
{
    "logging": {
        "level": "debug",
        "json": "true"
    }
}
```

# Usage

For help on individual commands, add --help following the command name.

When run with no arguments (or with -h/-help), `atlas` prints an usage message.


```bash
atlas <command> <subcommand> [flags]
```

# Available Commands

* [kms](./kms.md)
* [version](./version.md)
* [test](./test.md)
