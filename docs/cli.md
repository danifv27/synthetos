# Command-Line Interface

"Synthetos" is derived from the Greek word "synthetos" (συνθετος), which means "put together," "combined," or "composed." In the context of creating something synthetic or artificial, it could be interpreted as "constructed," "fabricated," or "built." 
It supports several commands, each accessible through the application binary. It can be configured in three different ways:

* Using a configuration file in Json format. This file should be placed in one of these paths: `/etc/synthetos.json`, `$HOME/.synthetos.json` or `<BINARY PATH>/.synthetos.json`
* Through commands flags.
* Via environment variables

ℹ️ command flags have precedence over environment values

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

When run with no arguments (or with -h/-help), `synthetos` prints an usage message.


```bash
synthetos <command> <subcommand> [flags]
```

# Available Commands
* [test](./test.md)
* [version](./version.md)
