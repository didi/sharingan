### 回放模块配置

<br>

回放时，被测模块的基本信息 根据 [本地回放](../README.md#4本地回放) 和 非本地回放 的不同，分别读取自不同的来源。

* 本地回放：读取自本地配置文件 [conf/moduleinfo.json](../../../replayer-agent/conf/moduleinfo.json) 
* 非本地回放：读取自 [conf/app.toml](../../../replayer-agent/conf/app.toml) 下 [http_api]小节module_info字段 配置的http接口。接口设计详见[http接口说明](./http_api.md#6-module_info)

<br>

下面主要针对本地回放时 conf/moduleinfo.json 内的配置字段进行解释说明。

##### 1. 配置示例
```json
{
  "data": [
    {
      "name":"example",
      "data":"[{\"key\":\"listen-addr\",\"value\":\"127.0.0.1:9999\"}]"
    }
  ]
}
```
<br>

##### 2. 配置说明

| moduleinfo.json字段说明 |  |
| :-----| :-----|
| data | []Module, Module类型的数组，数组每个值代表一个模块 |

| Module类型说明 | 类型 | 说明 |
| :-----| :----- | :----- |
| name | string | 模块名 |
| data | string | []KV类型的json串，存储模块详细信息 |

> 温馨提示：
>
> name尽量与编译后的$binName保持一致。如果存在冲突，可以增加前缀'\*-'，即\*-$binName。
>
> 如果name形如'\*a-b-c'，则尽量保证c具有可识别性，因为，[覆盖率统计回放](../replayer-codecov.md#1-覆盖率报告)会通过 *c* 来获取SUT进程信息并重启SUT。

| KV类型说明 | 类型 | 说明 |
| :-----| :----- | :----- |
| key | string | 与模块相关的一些信息，如监听地址listen-addr等 |
| value | string | 与key对应的具体值 |

本地回放时，KV类型中key只需支持一个"listen-addr"即可：
* listen-addr：必选，SUT的监听地址，一般为127.0.0.1:xxxx。

* department：可选，模块所属部门，默认空(则为default部门)。非空时，同时流量配置的读取自ES，则 会按部门字段读取es_url地址。es_url配置可详见：[回放Agent配置](../replayer-conf.md#5-es_url)
