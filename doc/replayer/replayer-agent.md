### Replayer-Agent一键安装和启动脚本

<br>

[./replayer-agent/control.sh](../../replayer-agent/control.sh) 脚本是 [流量回放-1. 配置并启动Replayer-Agent](./README.md#1-配置并启动replayer-agent) 的一键安装和启动脚本。 

* go mod方式脚本：[./replayer-agent/control_gomod.sh](../../replayer-agent/control_gomod.sh) 。对于没有GO环境的机器，会自动安装golang1.13，并设置GOROOT、GOPATH。
* glide方式脚本：[./replayer-agent/control_glide.sh](../../replayer-agent/control_glide.sh) 。对于没有GO环境的机器，会自动安装golang1.10、glide，并设置GOROOT、GOPATH。

control.sh 默认使用 go mod方式，即control_gomod.sh 脚本。

<br>

用户可以 **基于此脚本** 进行修改和定制。

##### 1. 环境

* GO 「任意版本官方golang，或者定制版本golang都可以」
* Glide 「低版本GO包管理工具」
* Go mod 「高版本GO包管理工具，go原生自带」

GO>=1.13推荐go mod, 低版本推荐glide

<br>

##### 2. 使用

下面的命令都是基于 **sharingan/replayer-agent** 目录操作的。

> 构建（build）

检测GO环境并编译Replayer-Agent服务。

```shell
sh control.sh build
```
如果提示如下错误：
```text
curl: (7) Failed to connect to raw.githubusercontent.com port 443: Connection refused
```
可能是raw.githubusercontent.com域名不通，建议配置个代理；或者 配置hosts (151.101.108.133 raw.githubusercontent.com)；再或者 自己安装go后，重新执行脚本构建。可参考: [go安装](https://github.com/golang/go#download-and-install)

<br>

> 启动（start）

启动Replayer-Agent服务。
```shell
sh control.sh start
```

<br>

> 停止（stop）

停止Replayer-Agent服务。
```shell
sh control.sh stop
```

<br>

> 重启（reload）

重启Replayer-Agent服务。
```shell
sh control.sh reload
```
