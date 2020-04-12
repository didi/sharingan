# 流量录制 - 压测

## 一、压测环境

* 压测机器：CPU: 4 & 内存: 16G & 硬盘大小: 100G ，docker机器
* 压测工具：[hey](https://github.com/rakyll/hey)，同样配置的docker机器远程压测
* 压测参数：hey -m POST -z 30s -c 1000 【持续30s，并发1000个请求】
* golang版本：go1.10.8

## 二、压测详情

分别压测五次避免环境等因素干扰，加录制后，QPS下降3%左右(2.4w左右 → 2.32w左右)。

| 正常http服务（不加录制代码） | 正常http服务（添加录制代码） |
| --- | --- |
| Requests/sec: 23987.6205 | Requests/sec: 23749.7493 |
| Requests/sec: 24281.8419 | Requests/sec: 24500.1881 |
| Requests/sec: 24435.5325 | Requests/sec: 23319.1393 |
| Requests/sec: 24099.6204 | Requests/sec: 23459.4222 |
| Requests/sec: 24452.9943 | Requests/sec: 21446.2462 |

## 三、压测代码

docker机器获取cpu核数有问题，我们都显示进行了设置：runtime.GOMAXPROCS(4)，[参考](https://mp.weixin.qq.com/s/rDjTqqR0q4VTSQrYFzbR7w)

### 3.1、正常http服务（不加录制代码）

``` go
package main

import (
        "fmt"
        "log"
        "net/http"
        "runtime"
        _ "net/http/pprof"
)

func main() {
        runtime.GOMAXPROCS(4)
        go http.ListenAndServe(":9981", nil)
        http.HandleFunc("/", indexHandle)
        log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello world!\n")
}
```

### 3.2、正常http服务（添加录制代码）

``` go
package main

import (
        "fmt"
        "log"
        "net/http"
        "runtime"
        _ "net/http/pprof"

        _ "github.com/didichuxing/sharingan/recorder"
)

func main() {
        runtime.GOMAXPROCS(4)
        go http.ListenAndServe(":9981", nil)
        http.HandleFunc("/", indexHandle)
        log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello world!\n")
}

```
