# AutoRun
自己实现一个自动检测文件变化并且进行重新编译运行的应用.

## Useage
使用 [goenv](https://bitbucket.org/ymotongpoo/goenv) 初始化好 GOPATH 等等环境变量, 然后在当前项目目录运行 autorun 即可,
其会自动根据当前目录设置为编译后的项目名称进行 watch -> build -> start

## Deps
go get github.com/howeyc/fsnotify