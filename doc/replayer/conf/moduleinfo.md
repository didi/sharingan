### 回放模块配置

回放时，被测模块的基本信息 根据 [本地回放](../README.md#4本地回放) 和 非本地回放 的不同，分别读取自不同的来源。

* 本地回放：读取自本地配置文件 [conf/moduleinfo.json](../../../replayer-agent/conf/moduleinfo.json) 
* 非本地回放：读取自 [conf/app.toml](../../../replayer-agent/conf/app.toml) 下 [http_api]小节module_info字段 配置的http接口。接口设计详见[http接口说明](./http_api.md#6-module_info)


下面主要针对 conf/moduleinfo.json 内的配置字段进行解释说明。

##### 1. 配置示例
```json
{
  "data": [
    {
      "name":"example",
      "data":"[{\"key\":\"context\",\"value\":\"localhost\"},{\"key\":\"listen-addr\",\"value\":\"127.0.0.1:9999\"},{\"key\":\"department\",\"value\":\"Biz\"}]"
    }
  ]
}
```

##### 2. 配置说明

| moduleinfo.json字段说明 |  |
| :-----| :-----|
| data | []Module, Module类型的数组 |

| Module类型说明 | 类型 | 说明 |
| :-----| :----- | :----- |
| name | string | 模块名 |
| data | string | []KV类型的json串 |

| KV类型说明 | 类型 | 说明 |
| :-----| :----- | :----- |
| key | string | 与模块相关的一些数据，如监听地址listen-addr等 |
| value | string | 与key对应的value值 |

目前KV类型中key支持的具体值有：
* context：必选，录制机器的地址，与录制流量的Context字段对应，以区分不同模块。一般一个模块找一台机器录制流量即可。

* listen-addr：必选，SUT的监听地址，一般为127.0.0.1:xxxx。

* department：可选，模块所属部门，默认空(则为default部门)。若非空，在非本地回放时，会按部门字段读取es_url地址。es_url配置可详见：[回放Agent配置](../replayer-conf.md#5-es_url)
