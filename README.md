# todushka

The simplest possible project, the main goal of which is to develop the skill of writing code.

For details, see the [Wiki](https://github.com/jtprogru/todushka/wiki) of the project.

By default, the configuration file is searched by the path `$HOME/.todushka/config.yaml`. Config example:

```yaml
---
server:
  addr: ":8082"
  log_level: "trace"

db:
  host: "127.0.0.1"
  port: "6432"
  username: "postgres"
  password: "postgres"
  db_name: "todushka"
  ssl_mode: "disable"
```

## License

[WTFPL](http://www.wtfpl.net)
