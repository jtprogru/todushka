# todushka

Максимально простой проект, главной целью которого является отработка навыка написания кода.

Подробности смотреть в [Wiki](https://github.com/jtprogru/todushka/wiki) проекта.

Конфигурационный файл ищется по пути `$HOME/.todushka/config.yaml`. Пример конфига:

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
