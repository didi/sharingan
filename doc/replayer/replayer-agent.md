### Agent一键安装和启动脚本

<br>

[./replayer-agent/control.sh](../../replayer-agent/control.sh) 脚本是 [流量回放-1. 配置并启动Agent](./README.md#1-配置并启动agent) 的一键安装和启动脚本。 

<br>

用户可以 **基于此脚本** 进行修改和定制。

##### 1. 环境

* GO 「任意版本官方golang，或者定制版本golang都可以」
* Glide 「低版本GO包管理工具」
* Go mod 「高版本GO包管理工具，GO>=1.13原生自带」

<br>

##### 2. 使用

下面的命令都是基于 **sharingan/replayer-agent** 操作的。

> 构建（build）

检测GO环境并编译Agent服务。

此脚本默认使用glide包管理工具，对于没有GO环境的机器，会自动安装golang1.10、glide，并设置GOROOT、GOPATH。 [golang安装](https://github.com/didichuxing/sharingan-go)。

使用gomod包管理的脚本即将提供~

```shell
sh control.sh build
```

> 启动（start）

启动Agent服务。
```shell
sh control.sh start
```

> 停止（stop）

停止Agent服务。
```shell
sh control.sh stop
```

> 重启（reload）

重启Agent服务。
```shell
sh control.sh reload
```
