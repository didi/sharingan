# 流量回放实现原理

Golang的流量回放 主要基于 [sharingan/replayer](../../replayer) 包 及 定制版的golang 实现。

相比Golang的流量录制原理，流量回放原理更简洁清晰一些。看下面的原理前，请先熟悉要用到的 [名词解释](./README.md#一名词解释)

## 一、简介

流量回放的前提是基于录制的流量进行操作。

如果录制的流量只有Inbound请求，没有Outbound请求，那么，回放过程非常简单。只需构造http请求发给SUT，等待SUT返回Response后进行对比即可。

但实际业务中，不仅会有Outbound请求，而且Outbound请求还很多，协议也各种各样。

因此，流量回放首要解决的问题有：
 1. 如何拦截SUT的Outbound请求，将其转发给Replayer-Agent的Mock Server。
 2. 如何在录制的流量里，选择最合适的Outbound返回给SUT。

## 二、Outbound请求拦截点
看过 [流量录制拦截点选择](https://github.com/didi/sharingan/wiki/%E6%8B%A6%E6%88%AA%E7%82%B9%E9%80%89%E6%8B%A9) 可知，golang的录制是在语言标准库层面做的拦截。

同理，流量回放也是在语言标准库层面做的拦截。修改系统调用 [syscall.Connect](https://github.com/golang/go/blob/release-branch.go1.10/src/syscall/syscall_unix.go#L225) 方法，将原本的socket地址sa替换为Mock Server地址。
```shell script
func Connect(fd int, sa Sockaddr) (err error) {
	ptr, n, err := sa.sockaddr()
	if err != nil {
		return err
	}
	return connect(fd, ptr, n)
}
```
实现上面修改的正是 [sharingan/replayer](../../replayer) 包，基于开源 [gomonkey](https://github.com/agiledragon/gomonkey) 库mock syscall.Connect 方法，解决了回放的第一个问题。

## 三、回放剧本传递

[sharingan/replayer](../../replayer) 包拦截了SUT的Outbound请求，将其转发给Replayer-Agent的Mock Server。

![replay-theory](http://img-hxy021.didistatic.com/static/sharingan/replay_theory.png)

如上图，回放剧本的传递过程如下：
  1. 用户浏览Web Server首页(:8998)，筛选一个流量，点击回放
  2. Web Server根据流量的Inbound Request，构造HTTP Request，发送给SUT
  3. SUT若不依赖其他下游，则直接返回HTTP Response给Web Server，跳到第8步。
  4. SUT若依赖其他下游，则replayer包会将下游请求重定向到Mock Server(:3515)。
  5. Mock Server接收到SUT的请求后，匹配最合适的Outbound Request，并根据剧本，返回对应的Outbound Response给SUT
  6. SUT接收到Outbound Response后，进行后续逻辑处理；若SUT依赖多个下游，则重复4，5步骤。
  7. SUT最后返回HTTP Response给Web Server
  8. Web Server收到SUT的HTTP Response后，与剧本的Inbound Response对比，给出回放结果。

上面过程完成了一次单流量回放。对于回放并发度大于1的情况，Mock Server如何识别接收到SUT的请求属于哪个流量呢？请详见：[并发回放实现](#五并发回放实现)

## 四、Outbound请求匹配

Mock Server有个非常重要的工作，就是匹配Outbound请求，这直接决定着回放的精确度和成功率。

### 1. 匹配算法
理论上来说，一个程序执行的过程中时间顺序是确定的，通过录制的时序就可以做回放。但是在现实的场景中，这种实现非常脆弱且不满足需求。因为大部分重构都需要调整调用顺序，如果回放完全基于调用顺序，则无法满足重构的功能回归的需求。

为此，Mock Server使用类似**全文检索**的模式，通过 **分词+打分**，实现如下匹配算法：

* n-gram 分词：n 取值是 16 个 byte
* 按 phrase 模式进行分词后的匹配打分
* 分词后的首个 chunk 进行加权匹配，因为 http 头部的 url 最有区分度
* 首个 chunk 要求是连续16个 readable bytes，过滤掉了二进制内容。主要是匹配 thrift 请求过滤到 size 的差异引起的问题
* 优先按顺序向后搜索（阈值 0.85），如果第一轮匹配不上则回到头部从头搜索（阈值0.1）

下面简化下匹配算法核心步骤：

![replay-match](http://img-hxy021.didistatic.com/static/sharingan/replay_match.png)

a) 匹配当前请求时，若上一次匹配成功，则从上一次匹配成功的请求（lastMatchedIndex）的下一个请求开始匹配，否则就从第一个请求开始匹配；

b) 将请求切分成长度为16字节的数组，依次取16字节与所有Outbound请求进行匹配，若请求里包含该16字节数组，则权重加1；

c) 判断权重最高的Outbound请求是否达标（权重是否超过阈值）。达标则匹配成功，否则匹配失败。

### 2. 优化算法
在实践过程中，发现 对于相似Outbound请求比较多的流量，匹配重复率较高。因为当中间一个请求匹配失败后，后面所有的请求都会重头开始匹配。最终导致多个线下请求匹配同一个线上请求，进而导致剩余的线上请求报missing错误。

因此，基于上面的匹配算法，增如下两点优化：
 1. 新增一个全局游标MaxMatchedIndex，记录当前已匹配请求的最大下标，当新请求到来时，优先从该游标之后的线上请求选择匹配：如果游标前后同时出现权重最高的线上请求，则选择游标之后的线上请求作为算法的匹配结果。
 2. 新增一个降权系数，当线上请求已被之前的请求匹配过之后，在本次请求的匹配过程中，它的匹配权重会降低为：权重*降权系数。
 
在实际使用过程中，降权系数选择为90%时，匹配效果非常好，成功率相当高。

## 五、并发回放实现

通过 [回放剧本传递](#三回放剧本传递) 可知，mock库解决了并发度=1的回放过程。

为了提高RD和QA的测试效率，需要支持并发度>1的并发回放。

### 1. 并发原理

![replay_parallel](http://img-hxy021.didistatic.com/static/sharingan/replay_parallel_v2.png)

如上图，基本思路如下：

通过sessionID关联 Web Server、SUT Server、Mock Server，使Mock Server能够识别从SUT接收的请求属于哪个流量。

 1. Web Server → SUT Server 【构造HTTP Request，传递sessionID】

 2. SUT Server维持sessionID 【Mock TCP连接的[Read方法](https://github.com/golang/go/blob/release-branch.go1.10/src/net/net.go#L172)，读取sessionID】

 3. SUT Server → Mock Server【Mock TCP连接的[Write方法](https://github.com/golang/go/blob/release-branch.go1.10/src/net/net.go#L184)，传递sessionID】
 
```shell script
// Read implements the Conn Read method.
func (c *conn) Read(b []byte) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	n, err := c.fd.Read(b)
	if err != nil && err != io.EOF {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, err
}

// Write implements the Conn Write method.
func (c *conn) Write(b []byte) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	n, err := c.fd.Write(b)
	if err != nil {
		err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, err
}
```
[sharingan/replayer](../../replayer) 包通过mock上面两个方法，来读取和传递sessionID。

### 2. SUT维持sessionID

SUT每接收Web Server的一个http请求，都会开启一个goroutine，同时，一般都会在这同一个goroutine内完成下游请求的调用，最终将http响应返回给Web Server。所以，针对这种情况，在同一个goroutine内维护sessionID是很简单的事。

但 对于通过新起一个goroutine来调用下游请求的SUT，维护sessionID需要基于录制里讲到的 [链路追踪](https://github.com/didi/sharingan/wiki/%E9%93%BE%E8%B7%AF%E8%BF%BD%E8%B8%AA) 原理，基于定制版的golang实现。

```shell script
// GetCurrentGoRoutineId get RoutineId in case delegate
func GetCurrentGoRoutineId() int64 {
    _g_ := getg()
    if _g_.delegatedFromGoid != 0 {
        return _g_.delegatedFromGoid
    }
    return _g_.goid
}
```
通过上面的GetCurrentGoRoutineId方法，使用goroutineID来关联inbound和outbound请求，以实现sessionID的维护。

### 3. 并发优化

上面的方案已经可以很好的实现并发回放，但并不支持同一sessionID的流量并发回放。

因此，增加如下一点优化：

* 用traceID代替sessionID，支持同一sessionID的流量并发回放

## 六、时间回放原理

流量回放是将过去发生的流量在当下进行回放，对于那些对时间敏感的流量，回放失败率很高。

为了能实现将 回放时间 倒回到 录制时间，参考[并发回放](#五并发回放实现)传递sessionID的原理，回放时Web Server将录制时间戳传递给SUT服务。

```shell script
// Now returns the current local time.
func Now() Time {
	sec, nsec, mono := now()
	sec += unixToInternal - minWall
	if uint64(sec)>>33 != 0 {
		return Time{uint64(nsec), sec + minWall, Local}
	}
	return Time{hasMonotonic | uint64(sec)<<nsecShift | uint64(nsec), mono, Local}
}
```

[sharingan/replayer](../../replayer) 包mock上面的 [time.Now](https://github.com/golang/go/blob/release-branch.go1.10/src/time/time.go#L1043) 方法，以实现时间的回放。
