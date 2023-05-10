# atlas version

The version command provides information about the current version of the software being used. It displays information such as the version number, build number, release date, and other relevant information. In addition, it may also provide information about the underlying operating system or platform that the software is running on.

Although it's not a standalone command, it is include in all the binaries created using this framework.

## Options
| Flag                 | Environment Variable      | Default Value | Description |
| :--------------------| :-------------------------| :------------ | :---------- |
| --version.output \<string> | SC\_VERSION\_OUTPUT | pretty | Specify the output format. Valid options are text (default) or json. (pretty\|json). |

### Config file

```json
{
    "atlas": {
        "version": {
            "output": "json"
        }
    }
}
```