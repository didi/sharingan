# 流量录制 - agent

功能：接受录制的流量，进行流量筛选、比例控制等，最终存入本地日志文件

## 一、快速开始

### 1、编译（build）

```shell
# 定制版或者官方golang都可以，需要1.9+版本
export GOROOT=/path/to/your/GOROOT
export PATH=$GOROOT/bin:$PATH

# 默认使用gomod方式进行包管理，其它方式参考实现
cd /path/to/your/sharingan/recorder-agent
sh control.sh build
```

### 2、启动（start）

```shell
cd /path/to/your/sharingan/recorder-agent
sh control.sh start
```

### 3、停止（stop）

```shell
cd /path/to/your/sharingan/recorder-agent
sh control.sh stop
```

### 4、重启（reload）

```shell
cd /path/to/your/sharingan/recorder-agent
sh control.sh reload
```

## 二、日志收集

* 默认recorder-agent会将流量以日志的形式在本地保存（见log/recorder.log）
* 默认流量日志按照小时切割，保留最近4个小时的数据，可以配置。（见conf/app.toml）
* 需要配置日志收集入ES，这样更方便使用（经典路线：log -> kafka -> ES）

### 2.1、日志收集入ES

* 将录制日志存到ES，日志示例

``` js
{"Context":"xxxxxxxx ","SessionId":"1585644498958060856-9","ThreadId":9,"TraceHeader":null,"NextSessionId":"1585644498959300110-9","CallFromInbound":{"ActionIndex":0,"OccurredAt":1585644498958075085,"ActionType":"CallFromInbound","Peer":{"IP":"::1","Port":52844,"Zone":""},"UnixAddr":{"Name":"","Net":""},"Request":"GET / HTTP/1.1\r\nHost: localhost:9999\r\nUser-Agent: curl/7.54.0\r\nAccept: */*\r\n\r\n"},"ReturnInbound":{"ActionIndex":1,"OccurredAt":1585644498958953830,"ActionType":"ReturnInbound","Response":"HTTP/1.1 200 OK\r\nDate: Tue, 31 Mar 2020 08:48:18 GMT\r\nContent-Length: 13\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello world!\n"},"Actions":[{"ActionIndex":0,"OccurredAt":1585644498958681859,"ActionType":"CallOutbound","SocketFD":8,"Peer":{"IP":"127.0.0.1","Port":8888,"Zone":""},"ResponseTime":1585644498958887823,"UnixAddr":{"Name":"","Net":""},"Request":"GET / HTTP/1.1\r\nHost: 127.0.0.1:8888\r\nUser-Agent: Go-http-client/1.1\r\nAccept-Encoding: gzip\r\n\r\n","Response":"HTTP/1.1 200 OK\r\nDate: Tue, 31 Mar 2020 08:48:18 GMT\r\nContent-Length: 12\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello test!\n","CSpanId":""},{"ActionIndex":1,"OccurredAt":1585644498958953830,"ActionType":"ReturnInbound","Response":"HTTP/1.1 200 OK\r\nDate: Tue, 31 Mar 2020 08:48:18 GMT\r\nContent-Length: 13\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello world!\n"}],"TraceId":"","SpanId":""}
```

* 索引mapping映射

``` js
{"Context":{"type":"text"},"TraceHeader":{"index":false,"type":"keyword"},"ThreadId":{"type":"long"},"Actions":{"properties":{"Response":{"type":"text"},"UnixAddr":{"dynamic":"false","type":"object","enabled":false},"ActionType":{"type":"keyword"},"CSpanId":{"type":"keyword"},"OccurredAt":{"type":"long"},"Content":{"type":"text"},"FileName":{"type":"text"},"Peer":{"type":"object","enabled":false},"Request":{"type":"text","doc_values":false},"SocketFD":{"type":"long"},"ActionIndex":{"type":"long"},"ResponseTime":{"type":"long"}}},"ReturnInbound":{"properties":{"Response":{"type":"text"},"ActionType":{"type":"keyword"},"OccurredAt":{"type":"long"},"ActionIndex":{"type":"long"}}},"NextSessionId":{"type":"keyword"},"TraceId":{"type":"keyword"},"message":{"index":false,"type":"keyword","doc_values":false},"CallFromInbound":{"properties":{"UnixAddr":{"dynamic":"false","type":"object","enabled":false},"ActionType":{"type":"keyword"},"OccurredAt":{"type":"long"},"Peer":{"type":"object","enabled":false},"Request":{"type":"text"},"ActionIndex":{"type":"long"}}},"SessionId":{"type":"keyword"},"SpanId":{"type":"long"}}
```

## 三、其它

### 1、端口设置

位置：/path/to/your/sharingan/recorder-agent/conf/app.toml
默认：9003
