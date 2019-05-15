# Reloader
**`DEPRECATED`**
减少文件修改后需要手动重启服务测试的流程，通过监控文件改动来自动重启，适用于不影响出错语法的内容修改项目。

## dependency

- `"github.com/fsnotify/fsnotify"`  based on epool
- `"github.com/guoruibiao/commands"`

## TODO
- [ ] file filter chain based on .gitignore.
- [ ] kill old process and reload new process.
- [ ] command line parameters parser.
