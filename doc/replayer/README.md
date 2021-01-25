# 流量回放

<br>

流量回放，即基于录制的流量在线下对改造后的代码进行回放，以验证线下代码逻辑的正确性。其中，流量录制可以根据需求，录制线上机器的真实流量，或者录制本地机器的测试流量。

<br>

## 一、名词解释
* **SUT**: System Under Test, 在这里特指被测业务服务。目前支持http server 和 grpc server。
* **Inbound**: 指SUT对外提供的 HTTP/Thrift 接口的请求和响应，即Inbound Request/Response。目前支持对外提供 HTTP/Thrift 接口的SUT回放。
* **Outbound**: 指SUT提供的 http/thrift 接口内调用的下游请求和响应，即Outboud Request/Response。目前支持的Outbound协议有 MYSQL、REDIS、HTTP、Binary Thrift、Compact Thrift。
* **流量**: 指在TCP层对SUT录制的Inbound请求/响应 和 对应的Outbound请求/响应。
* **回放**: 基于录制的一条流量，Replayer-Agent根据Inbound Request构造 HTTP/Thrift Request并对SUT发起请求；其中SUT的Outbound请求会发送到Mock Server进行流量匹配并返回Outbound响应；最后，Replayer-Agent将收到的SUT HTTP/Thrift Response与Inbound Response做对比，给出回放结果。
* **Replayer-Agent**: 包括回放过程的Web Server和Mock Server。其中Web Server默认监听8998端口，主要负责流量搜索，回放请求的构造和发起，以及回放结果的对比和展示。
* **Mock Server**: 默认监听**3515**端口，主要负责接收SUT的下游请求，然后与回放流量的Outbound请求匹配，将最匹配的Outbound响应返回SUT。
* **噪音**: 即在Inbound Response对比和Outbound Request匹配过程里，出现的不影响回放结果和匹配度的diff字段，如时间戳。Replayer-Agent回放结果页支持上报噪音，方便再次回放时精确回放结果。
* **DSL**: 特指ElasticSearch的查询DSL。Replayer-Agent首页支持上报流量查询DSL，方便后面直接复用。

<br>

## 二、回放接入

回放前提：已经完成流量录制。[录制接入文档](../recorder/README.md)

对于服务启动阶段有TCP请求的SUT，如初始化连接池等，推荐 服务启动顺序：
* 先启动Replayer-Agent 
* 再启动SUT

#### 1. 配置并启动Replayer-Agent

> 温馨提示：Replayer-Agent默认配置的是 **[本地回放](#4本地回放)**，即 仅依赖本地配置文件的回放。
* 如需 读取录制的线上流量，只需根据[Replayer-Agent配置](./replayer-conf.md#5-es_url)修改es_url字段即可；
* 如需 读取上报到自有服务的噪音或DSL，只需根据[Replayer-Agent配置](./replayer-conf.md#4-http_api)修改http_api字段即可；
* 如需 新增本地回放模块，请参考[本地回放接入](./replayer-local.md)。

##### 快速安装(推荐)
```shell script
# mac
wget https://github.com/didi/sharingan/releases/download/replayer-agent%2Fv1.3.0/replayer-agent.mac
# linux
wget https://github.com/didi/sharingan/releases/download/replayer-agent%2Fv1.3.0/replayer-agent.linux
```
最新版本请查看 [releases](https://github.com/didi/sharingan/releases) 

将上步wget的bin文件与 [replayer-agent/conf](https://github.com/didi/sharingan/releases/download/replayer-agent%2Fv1.3.0/conf.tar) 和 [replayer-agent/template](https://github.com/didi/sharingan/releases/download/replayer-agent%2Fv1.3.0/template.tar) 放到一个目录下，启动bin文件即可。

##### 编译安装
首先，确保本地存在go环境(官方或定制版均可，且版本无要求), 并根据本地go版本选择go mod或glide来安装sharingan的依赖。
```shell script
git clone https://github.com/didi/sharingan.git
&& cd sharingan && go mod download # 低版本go执行 glide install 
```
然后编译并启动Replayer-Agent：
```shell script
cd ./replayer-agent && go build && nohup ./replayer-agent >> run.log 2>&1 &
```
> Replayer-Agent一键安装和启动 [脚本](../../replayer-agent/control.sh) 及其 [使用方法](./replayer-agent.md)

<br>

#### 2. 配置并启动SUT

目前，SUT既支持http server，也支持 grpc server。

##### 2.1 http server

> 需要使用定制的golang，并通过指定tag来编译 引入回放包的SUT代码。

首先，配置定制版golang环境，目前支持go1.10 ~ go1.15，参考：[golang安装](https://github.com/didi/sharingan-go/tree/recorder)
```shell script
curl https://raw.githubusercontent.com/didi/sharingan-go/recorder/install/go1.13 | sh
&& export GOROOT=/tmp/recorder-go1.13
&& export PATH=$GOROOT/bin:$PATH
```

然后，修改SUT的main.go代码，引入sharingan包。参考：[example](../../example/replayer/main.go)
```go
import _ "github.com/didi/sharingan"
```

最后，编译并启动SUT服务。
```shell script
go build -tags="replayer" -gcflags="all=-N -l"
&& nohup ./$project >> run.log 2>&1 &
```
> SUT一键接入和启动 [脚本](../../example/replayer/sut_replayer.sh) 及其 [使用方法](./replayer-sut.md)

至此，浏览器打开 [http://127.0.0.1:8998](http://127.0.0.1:8998) 或 local_ip 即可开始回放啦~

##### 2.2 grpc server

回放使用的 [grpc server](../../grpc-server) 基于开源的 [grpc-go](https://github.com/grpc/grpc-go) 进行了定制化，目前支持的版本是v1.33.2。

引入定制的 [grpc server](../../grpc-server) 方式有两种：

方式一：

将 [grpc server](../../grpc-server) 代码提交到SUT的代码仓库，直接使用。

方式二：

将 [grpc server](../../grpc-server) 提交到一个新的代码仓库后，发包v1.33.2，然后在SUT的 go.mod 内使用replace 替换开源的grpc包:

> replace google.golang.org/grpc => xxx/grpc v1.33.2

<br>

回放其他操作，同 [2.1 http server](#2.1-http-server) 。

不同的是，模块配置需增加 server-type 字段，且赋值为 GRPC_SERVER。详见：[回放模块配置](./conf/moduleinfo.md)

<br>

## 三、使用手册

按照上面的接入流程逐步操作 并 成功启动SUT和Replayer-Agent后，即可开始回放之旅~

首页点击"流量查询"搜索流量后，即可点击"运行"开始回放。
![web_index](http://img-hxy021.didistatic.com/static/sharingan/web_index_v2.png)

#### 1.指南

为了大家能够轻松、愉悦地开始回放过程，特意整理了回放过程中常见的使用技巧及指南，方便大家进一步理解流量回放。

[流量 搜索指南](./guide/search.md)

[流量 回放指南](./guide/replay.md)

[噪音/DSL 上报指南](./guide/report.md)

#### 2.覆盖率统计

流量回放不仅支持上面[接入流程-配置并启动SUT](#2-配置并启动sut)展示的普通回放接入，还支持覆盖率统计方式的回放接入。

[覆盖率统计回放接入](./replayer-codecov.md)

#### 3.批量并发回放

流量回放不仅支持单个流量回放，还支持多个流量批量并发的回放。

[批量并发回放](./replayer-parallel.md)

#### 4.本地回放

本地回放 即 仅依赖本地配置文件的回放。 流量 读取本机录制的测试流量，噪音和DSL 读取上报到本机的噪音和DSL，模块信息 读取自本机配置文件。

[本地回放接入](./replayer-local.md)

Replayer-Agent默认配置的本地回放，如需修改，请参考：[Replayer-Agent配置](./replayer-conf.md#4-http_api)

#### 5.Replayer-Agent和SUT分开部署

如上 二-回放接入，默认将Replayer-Agent和SUT部署在一台机器，其实，Replayer-Agent和SUT是可以分开部署的。

* Replayer-Agent启动前，需要修改 [conf/moduleinfo.json](../../replayer-agent/conf/moduleinfo.json) 内SUT监听地址。
* 启动SUT时，需设置环境变量 REPLAYER_MOCK_IP 为Replayer-Agent的ip地址即可。

详见: [SUT启动脚本-3. 与Replayer-Agent分开部署](./replayer-sut.md#3-与replayer-agent分开部署)

#### 6.跨语言流量回放

新版本 replayer-agent/v1.2.0 支持跨语言流量回放。即 支持使用录制的PHP流量在go模块进行回放，来支持使用GO重构PHP模块的测试需求。

PHP流量特点：inbound是fastcgi协议，outbound支持http、Redis、MySQL、thrift协议。

GO流量特点：inbound是http/thrift协议，outbound支持http、Redis、MySQL、thrift协议。

新版本的 replayer-agent 会自动识别inbound协议，将fastcgi协议转为http协议，进行流量回放；同时会模拟mysql部分非业务交互，解决PHP流量未录制部分MySQL交互请求的问题。

<br>

## 四、常见问题
[常见问题及排查](./guide/troubleshoot.md)

[交流群](./guide/troubleshoot.md#交流群)

<br>

## 五、回放原理
[回放原理详解](./replayer-theory.md)
