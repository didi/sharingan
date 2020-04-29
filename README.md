# Sharingan

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/didi/sharingan/blob/master/LICENSE)

Sharingan，中文名：写轮眼，是漫画《火影忍者》中的一种瞳术，具有复制、幻术等能力，在幻术世界里，一切因素包括时间，地点，质量等，都由施术者掌控。

## 一、简介

Sharingan是一个基于golang的流量录制回放工具，录制线上真实请求流量进行回放测试，适合项目重构、回归测试等。

### 1.1、背景

随着微服务架构的兴起，服务之间的依赖关系变的越来越复杂，软件测试也面临新的挑战：系统升级频繁、服务依赖众多等等。

* 常见的测试方案（如：单元测试、系统测试等）构造和维护测试用例成本高，特别是业务复杂的场景。「构造测试数据**难**」
* 依赖第三方服务众多，线下测试环境不太稳定，经常出现下游服务不可用导致测试失败的情况发生。「维护测试环境成本**高**」

为此，我们需要开发一套工具来缓解上述问题。

### 1.2、方案

* 录制线上服务真实请求流量（包括下游调用流量），在线下进行回放，解决构造测试数据难的问题。「复制能力」
* 回放的时候匹配Mock下游调用，不再依赖具体的下游服务，解决维护测试环境成本高的问题。「幻术能力」

### 1.3、特性

* 支持下游流量录制。相比[tcpcopy](https://github.com/session-replay-tools/tcpcopy)、[goreplay](https://github.com/buger/goreplay)等方案，回放不依赖下游环境。
* 支持并发流量录制和回放。录制对服务影响小，回放速度更快。
* 支持时间重置、噪音去除、批量回放、覆盖率报告、常见协议解析等等。
* 支持写流量回放，不会污染应用数据。
* 不依赖业务框架，低应用浸入。

## 二、快速开始

### 2.1、使用示例

```shell
# Step1: 下载sharingan项目
$ git clone https://github.com/didi/sharingan.git
$ cd sharingan

# Step2: 使用定制版golang，以go1.13为例
$ sh install.sh go1.13 # 支持go1.10 ~ go1.14，限mac、linux amd64系统
$ export GOROOT=/tmp/recorder-go1.13
$ export PATH=$GOROOT/bin:$PATH

# Step3: 编译、后台启动replayer-agent「会占用3515、8998端口」
$ cd replayer-agent
$ go build
$ nohup ./replayer-agent >> run.log 2>&1 &

# Step4: 编译、后台启动example示例「会占用9999端口」
$ cd ../example
$ go build -tags="replayer" -gcflags="all=-N -l"
$ nohup ./example >> run.log 2>&1 &

# Step5: 打开回放页面
$ 浏览器打开，http://127.0.0.1:8998 # 非本机替换IP即可
$ 页面选择要回放的流量点执行          # 内置提前录制好的3条example示例流量
```

### 2.2、接入文档

* [录制接入文档](./doc/recorder/README.md) 「先录制」
* [回放接入文档](./doc/replayer/README.md) 「后回放」

## 三、技术方案

### 3.1、模块划分

* recorder: 流量录制包，录制流量本地文件存储、发送流量到录制agent等。
* recorder-agent：流量录制agent，单独进程启动，控制录制比例、流量存储等。
* replayer: 流量回放包，重定向连接到Mock Server、Mock时间、添加流量标识等。
* replayer-agent：流量回放agent，单独进程启动，查询流量、查询/上报噪音、流量diff、批量回放、生成覆盖率报告等。

### 3.2、整体架构图

<img width="725" height="521" alt="架构图" src="http://img-hxy021.didistatic.com/static/sharingan/arch_v2.png"/>

### 3.3、录制方案

* 修改golang源码，对外暴露Hook接口。「所有改动通过官方测试用例」
* 提供API串联不同goroutine之间的网络调用。「常见的http、mysql、redis流量都不需要特别设置」
* 提供单独的agent筛选流量、控制比例。
* 更多参考：[流量录制实现原理](https://github.com/didi/sharingan/wiki/%E6%B5%81%E9%87%8F%E5%BD%95%E5%88%B6%E5%AE%9E%E7%8E%B0%E5%8E%9F%E7%90%86)。

### 3.4、回放方案

* 连接重定向：将服务所有Connect网络调用重定向到Mock Server。「安装replayer-agent时候自带」
* 流量匹配：Mock Server会根据服务真实的下游请求匹配一个返回。「mock下游调用」
* 时间重置：将程序执行时间回退到录制执行时刻，尽量避免时间因素带来的干扰。
* 噪音去除：提供API可以将已知的噪音流量去掉，如：traceID，每次请求本来就不一样。
* 常见协议解析：会解析http、mysql、redis、thrift等协议，方便diff对比。
* 更多参考：[流量回放实现原理](https://github.com/didi/sharingan/wiki/%E6%B5%81%E9%87%8F%E5%9B%9E%E6%94%BE%E5%AE%9E%E7%8E%B0%E5%8E%9F%E7%90%86)。

## 四、演进之路

关于流量录制和回放，在内部进行过多次探索，主要经历下面三个阶段：

### 4.1、月光宝盒（串行录制、串行回放）

录制：利用tcpdump录制流量，改造router层将请求串行化，利用时间来分割请求。

回放：利用iptables转发流量到mock服务，匹配请求并mock返回。支持时间重置、流量对比等。

不足：录制流量覆盖率低，一次只能录制一个请求。iptables转发，噪音干扰严重。

### 4.2、fastdev（并行录制、串行回放）

录制：改造golang源码，利用goroutine+工作委托技术串联区分请求。[链路追踪原理](https://github.com/didi/sharingan/wiki/%E9%93%BE%E8%B7%AF%E8%BF%BD%E8%B8%AA)

回放：利用monkey mock库对connect系统调用的重定向，转发流量到mock服务。支持Dashboard、噪音去除等。[monkey mock原理](https://bou.ke/blog/monkey-patching-in-go/)

不足：录制接口和实现混合，golang多版本支持困难。不支持并发回放，启动阶段流量无法代理会失败，定时任务流量干扰严重。

### 4.3、sharingan（并行录制、并行回放）

录制：接口和实现分离，golang源码改造部分只暴露接口，具体录制实现单独提供包支持。确保golang源码改动通过官方测试，支持1.10~1.14所有版本；优化录制服务性能。

回放：添加流量标识，支持并发回放；支持启动阶段流量代理；利用定制版golang，消除定时任务流量干扰；时间重置不再依赖本地文件，支持replayer-agent单独部署；支持常见协议解析。

## 五、效果展示

### 5.1、流量回放

#### 5.1.1、单个回放

<img width="870" height="370" alt="单个回放" src="http://img-hxy021.didistatic.com/static/sharingan/replayer_s_v2.png"/>

对于下游请求很多且复杂的情况，支持对下游协议进行筛选
<img width="870" height="285" alt="单个回放-协议刷选" src="http://img-hxy021.didistatic.com/static/sharingan/protocol_v2.png"/>

#### 5.1.2、批量回放

批量回放的并发度默认是10，可通过增加[-parallel](https://github.com/didi/sharingan/blob/master/replayer-agent/control.sh#L160)参数修改。
<img width="870" height="303" alt="批量回放" src="http://img-hxy021.didistatic.com/static/sharingan/replayer_p_v2.png"/>

### 5.2、覆盖率报告

#### 5.2.1、整体报告

覆盖率报告支持覆盖率结果累计，即支持 多次 单个回放和批量回放后，统一生成覆盖率结果。
<img width="870" height="365" alt="整体报告" src="http://img-hxy021.didistatic.com/static/sharingan/codeCover_v2.png"/>

#### 5.2.1、覆盖详情

<img width="870" height="415" alt="覆盖详情" src="http://img-hxy021.didistatic.com/static/sharingan/codeCover_detail_v2.png"/>

## 六、更多

### 6.1、如何贡献

欢迎大家参与进来，更多参考[Contribute](./CONTRIBUTING.md)。

### 6.2、许可

基于Apache-2.0协议进行分发和使用，更多参考[LICENSE](./LICENSE)。

### 6.3、成员

[hueng](https://github.com/hueng)、[yj20060714](https://github.com/yj20060714)、[qiaodandedidi](https://github.com/qiaodandedidi)、[bikong0411](https://github.com/bikong0411)、[plpan](https://github.com/plpan)、[fzl-yty](https://github.com/fzl-yty)。

### 6.4、感谢

特别感谢[TaoWen](https://github.com/taowen) ，流量录制和回放初版设计者，为后续开源奠定了很好的基础。

### 6.5、联系我们

欢迎加入QQ交流群「群号:417146801」一起交流~

![QQ](http://img-hxy021.didistatic.com/static/sharingan/QQ_v2.JPG)

Tips：联系QQ群主可拉入微信交流群，日常交流**以微信群为主**。
