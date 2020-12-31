# 流量录制

录制服务真实请求流量，为后续流量回放提供数据支持。

## 一、接入流程

下面 第1-4步骤 是http server的标准录制流程，第5步骤 是 grpc serser的录制流程。

### 1、使用定制版golang【必须】

* 目前支持go1.10 ~ go1.14，参考：[golang安装](https://github.com/didi/sharingan-go/tree/recorder)

```shell
# go1.13使用示例，github域名不太稳定，失败可以换成其它方式，参考golang安装
curl https://raw.githubusercontent.com/didi/sharingan-go/recorder/install/go1.13 | sh
export GOROOT=/tmp/recorder-go1.13
export PATH=$GOROOT/bin:$PATH
```

### 2、修改项目

#### 2.1、引入sharingan包【必须】

* 引入包要在业务包之前，保证流量到来之前已经初始化

```go
import _ "github.com/didi/sharingan"
```

* 参考：[example](https://github.com/didi/sharingan/blob/master/example/main.go)

#### 2.2、特殊设置【非必须】

* 背景：使用goroutine对外网络调用时，需要显示的传递goroutineID，否则链路无法串联起来，详见：[链路追踪](https://github.com/didi/sharingan/wiki/%E9%93%BE%E8%B7%AF%E8%BF%BD%E8%B8%AA)
* tip1：定时任务的流量不会录制，涉及到的代码不需要修改。「我们只录制对外http接口整个流程的流量」
* tip2：http请求主流程不等待结果的异步网络调用，不需要设置。「我们只录制http请求阶段确定的流量」
* tip3：常见的第三方包「http、redis、mysql、thrift等」，经测试都可以正常进行录制，不需要修改。

```go
import "github.com/didi/sharingan"
  
// 修改之前的代码
-   go remoteRead()
  
// 修改之后的代码
+   go func(delegatedID int64) {
+       sharingan.SetDelegatedFromGoRoutineID(delegatedID)
+       defer sharingan.SetDelegatedFromGoRoutineID(0)
+       remoteRead()
+   }(sharingan.GetCurrentGoRoutineID())
```

* 参考：[example](https://github.com/didi/sharingan/blob/master/example/recorder/main.go)

### 3、指定tag编译【必须】

```shell
cd /path/to/your/project
go build -tags="recorder"
```

### 4、录制

#### 4.1、线下流量录制「本地存储流量」

```shell
export RECORDER_ENABLED=true              // 开启录制
export RECORDER_TO_FILE=/tmp/recorder.log // 指定文件存储，确保目录存在；不设置的话会再控制台输出流量
cd /path/to/your/project && ./$project    // 使用上一步编译生成二进制文件启动项目
```

* 录制成功标志：指定文件/tmp/recorder.log存在流量，一条流量占一行。
* 线下流量回放，参考：[本地回放](https://github.com/didi/sharingan/blob/master/doc/replayer/replayer-local.md)。

#### 4.2、线上流量录制「流量发送给recorder-agent」

```shell
export RECORDER_ENABLED=true                     // 开启录制
export RECORDER_TO_AGENT="http://127.0.0.1:9003" // 指定agent地址，确保recorder-agent已经启动
cd /path/to/your/project && ./$project           // 使用上一步编译生成二进制文件启动项目
```

* 启动recorder-agent：[recorder-agent](https://github.com/didi/sharingan/blob/master/doc/recorder/recorder-agent.md)
* 录制成功标志：/path/to/your/sharingan/recorder-agent/log/recorder.log存在流量，一条流量占一行。
* 线上流量回放，参考：[流量回放](https://github.com/didi/sharingan/tree/master/doc/replayer)。

### 5、grpc server录制

#### 5.1、与http server标准录制流程区别

##### 5.1.1、编译前，使用定制的grpc替换开源grpc【必须】

录制使用的 [grpc server](../../grpc-server) 基于开源的 [grpc-go](https://github.com/grpc/grpc-go) 进行了定制化，目前支持的版本是v1.33.2。

引入定制的 [grpc server](../../grpc-server) 方式有两种：

方式一：

将 [grpc server](../../grpc-server) 代码提交到SUT的代码仓库，直接使用。

方式二：

将 [grpc server](../../grpc-server) 提交到一个新的代码仓库后，发包v1.33.2，然后在SUT的 go.mod 内使用replace 替换开源的grpc包:

> replace google.golang.org/grpc => xxx/grpc v1.33.2

##### 5.1.2、指定新的tag编译【必须】

```shell
cd /path/to/your/project
go build -tags="recorder_grpc"
```

注意，此处tag值是 **"recorder_grpc"**。

<br>

grpc server的其他录制操作，同 上面http server标准录制流程的1-4步骤 。

#### 5.2、流量demo

详见 [grpc server录制流量demo](./traffic_grpc) 。

#### 5.3、待支持功能

##### 5.3.1、Outbound 为 grpc协议 的录制

即 grpc client端录制【待定制的[grpc-server](../../grpc-server)和[koala_grpc](../../recorder/koala_grpc)支持】

##### 5.3.2、grpc协议 的回放

即 inbound 和 outbound为grpc协议的回放【待[replayer-agent](../../replayer-agent)支持】

>温馨提示：

>对于使用grpc-server做PHP转GO的模块，跨语言流量回放是支持的。

replayer-agent已经支持 [grpc server回放](../replayer/README.md#22-grpc-server)，回放时，replayer-agent自动识别inbound的fastcgi协议，并转为http协议后进行回放。

grpc server框架流量链路：http协议→ grpc-gateway→ grpc协议→grpc server.

因此，对于使用http协议，且基于grpc-gateway的模块，跨语言回放不受影响。

## 二、最佳实践【**线上流量录制，强烈推荐**】

### 1、编译「生成两个bin文件」

```shell
# 以项目test为例
app=test

# 使用官方GO编译，不添加任何tag【正常bin】
export GOROOT=/path/to/official_go
export PATH=$GOROOT/bin:$PATH
go build -o ${app}

# 使用定制版GO编译，添加recorder tag【录制bin】
export GOROOT=/path/to/special_go
export PATH=$GOROOT/bin:$PATH
go build -tags="recorder" -o ${app}-recorder
```

### 2、启动「指定机器开启录制，使用agent收集流量」

```shell
# 以项目test为例
app=test

# 将xxx机器名替换为待录制机器的hostname，通常只在一台机器录制
if [ `hostname`x = "xxx机器名"x ] ; then
    export RECORDER_ENABLED=true
    export RECORDER_TO_AGENT="http://127.0.0.1:9003"
    app=${app}-recorder
fi

# 启动服务
./${app}
```

## 三、录制原理

[录制原理详解](https://github.com/didi/sharingan/wiki/%E6%B5%81%E9%87%8F%E5%BD%95%E5%88%B6%E5%AE%9E%E7%8E%B0%E5%8E%9F%E7%90%86)

## 四、常见问题

### 1、对正常服务影响

* [压测详情](https://github.com/didi/sharingan/blob/master/doc/recorder/hey.md)
* 建议只在一台机器上开启录制，其它机器不受任何影响。【降低影响面，参考最佳实现】
* 建议添加pprof监控，观察服务健康状况，逐步调高录制机器权重。【有问题可以下线机器，或者在机器上用正常bin重启】

### 2、支持情况

* 不支持https、http2.0(包括grpc)流量录制。
* 支持其它常见协议，如：http1.1、thrift、mysql、redis、mongo等。
