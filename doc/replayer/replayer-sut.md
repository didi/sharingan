### SUT一键接入和启动脚本

<br>

sut_replayer.sh 脚本是 [流量回放-2. 配置并启动SUT](./README.md#2-配置并启动sut) 示例 example 的一键安装和启动脚本，目前支持如下两种方式： 

* go mod方式：[./example/replayer/sut_replayer_gomod.sh](../../example/replayer/sut_replayer_gomod.sh) 。对于没有GO环境的机器，会自动安装golang1.13，并设置GOROOT、GOPATH。

* glide方式： [./example/replayer/sut_replayer_glide.sh](../../example/replayer/sut_replayer_glide.sh) 。对于没有GO环境的机器，会自动安装golang1.10、glide，并设置GOROOT、GOPATH。

sut_replayer.sh 默认使用 **go mod方式**，即 sut_replayer_gomod.sh 脚本。

<br>

用户可以 **基于此脚本** 进行修改和定制。

##### 1. 环境

* GO 「定制版本golang，目前支持go1.10 ~ go1.14」
* 包管理工具由SUT决定 (GO>=1.13推荐go mod, 低版本推荐glide)

<br>

##### 2. 使用

下面的命令都是基于 **sharingan/example/replayer** 名录操作的，相当于SUT的根目录。

> 构建（build）

检测GO环境并编译SUT服务。

```shell
sh sut_replayer.sh build     //普通回放
sh sut_replayer.sh build cov //覆盖率回放
```
如果提示如下错误：
```text
curl: (7) Failed to connect to raw.githubusercontent.com port 443: Connection refused
```
可能是raw.githubusercontent.com域名不通，建议配置个代理；或者 配置hosts (151.101.108.133 raw.githubusercontent.com)；再或者 自己安装go后，重新执行脚本构建。可参考: [sharingan-go安装](https://github.com/didi/sharingan-go/tree/recorder)

<br>

> 启动（start）

启动SUT服务。
```shell
sh sut_replayer.sh start     //普通回放
sh sut_replayer.sh start cov //覆盖率回放
```
<br>

> 停止（stop）

停止SUT服务。
```shell
sh sut_replayer.sh stop     //普通回放
sh sut_replayer.sh stop cov //覆盖率回放
```
停止SUT时，覆盖率回放方式 会给出覆盖率报告 及 可以直接查看的html链接。覆盖率报告详细说明：[覆盖率报告](./replayer-codecov.md#1-覆盖率报告)
![shell_sut_cov_stop](http://img-hxy021.didistatic.com/static/sharingan/shell_sut_cov_stop.png)

<br>

> 重启（reload）

重启SUT服务。
```shell
sh sut_replayer.sh reload     //普通回放
sh sut_replayer.sh reload cov //覆盖率回放
```

<br>

##### 3. 与Replayer-Agent分开部署

SUT与Replayer-Agent服务可以在不同的机器分开部署，其中Replayer-Agent的Mock Server监听端口3515也可以修改。
> 分开部署

  1. 修改Replayer-Agent [conf/moduleinfo.json](../../replayer-agent/conf/moduleinfo.json) 配置文件内listen-addr字段值为SUT真实地址。字段详解: [回放模块配置](./conf/moduleinfo.md)
  2. 启动Replayer-Agent服务。参见: [Replayer-Agent启动脚本](./replayer-agent.md)
  3. 修改脚本 [./example/replayer/sut_replayer.sh](../../example/replayer/sut_replayer.sh) 里的 REPLAYER_MOCK_IP 环境变量，为Replayer-Agent的ip地址。
  4. 重启SUT服务即可。

> 修改Mock Server 3515端口

  1. 使用新端口 修改 [Replayer-Agent配置-outbound](./replayer-conf.md#3-outbound) 字段
  2. 重启Replayer-Agent服务。参见: [Replayer-Agent启动脚本](./replayer-agent.md)
  3. 修改脚本 [./example/replayer/sut_replayer.sh](../../example/replayer/sut_replayer.sh) 里的 REPLAYER_MOCK_PORT 环境变量，为Mock Server新端口。
  4. 重启SUT服务即可。
