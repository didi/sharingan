# 流量回放 - agent

功能：接受录制的流量，进行流量筛选、比例控制等，最终存入本地日志文件

## 一、接入文档

### 1、要求（require）

* GO >= 1.9 「任意官方golang，或者定制版本golang都可以」
* Glide     「其它包管理方式参考」

### 2、编译（build）

```shell
export GOROOT=/path/to/your/GOROOT
export GOPATH=/path/to/your/GOPATH
export PATH=$GOROOT/bin:$PATH

cd /path/to/your/sharingan
glide install

cd /path/to/your/sharingan/replayer-agent
sh control.sh build
```

### 3、启动（start）

```shell
cd /path/to/your/sharingan/replayer-agent
sh control.sh start
```

### 3、停止（stop）

```shell
cd /path/to/your/sharingan/replayer-agent
sh control.sh stop
```

### 4、重启（reload）

```shell
cd /path/to/your/sharingan/replayer-agent
sh control.sh reload
```

## 二、其它

[replayer-agent配置文件说明](./replayer-conf.md)

[TODO]
