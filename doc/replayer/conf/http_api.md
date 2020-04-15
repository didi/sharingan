### 噪音/DSL/模块_http接口说明

<br>

回放时，噪音、DSL 根据 [本地回放](../README.md#4本地回放) 和 非本地回放 的不同，分别上报到不同的地方；模块信息也会读取自不同的来源：[回放模块配置](./conf/moduleinfo.md)。

* 本地回放：噪音 上报到本机 [conf/noise/](../../../replayer-agent/conf/noise) 目录下，DSL 上报到本机 [conf/dsl/](../../../replayer-agent/conf/dsl) 目录下，模块信息读取自本地配置。
* 非本地回放：噪音、DSL 上报到 [conf/app.toml](../../../replayer-agent/conf/app.toml) 下 [http_api]小节 配置的http接口。模块信息也读取自该配置接口。

<br>

回放Agent默认配置的是本地回放；对于想将 噪音、DSL、模块信息存入数据库的同学，请仔细阅读下面接口说明。
> 温馨提示：
>
> 接口实现 只要字段名和字段类型符合接口说明即可，**至于字段的具体值无需担心，因为所有的值都会通过前端js处理后传给后端，比如dsl字段、noise字段等**。

##### 1. dsl_push

接口: http://{{your_domain}}/dsl

定义 DSL上报接口。上报DSL可以保存从ES进行流量查询的条件信息。若dsl_push为空，则上报到本机 [conf/dsl](../../../replayer-agent/conf/dsl) 下的配置文件。

在回放首页，对于查询条件复杂且后面需要继续使用的流量，可通过点击"保存查询"来存储查询条件，后面在"快速查询"处直接输入上报时的tag名，即可自动填充各查询条件，提升测试效率。
![save_dsl](../../images/save_dsl.png)

| 接口说明 |  |
| :-----| :-----|
| Method | POST |
| Content-Type | application/x-www-form-urlencoded |

| 参数 | 类型 | 说明 |
| :-----| :----- | :----- |
| project | string | 必选，被测模块名 |
| dsl | string | 必选，上报查询条件组成的json串 |
| tag | string | 必选，给该dsl自定义名字，方便后面查询 |
| user | string | 可选，上报用户名 |

| 返回说明 |  |  |
| :-----| :-----| :-----|
| errno | int | 错误码，0 成功，非0 失败 |
| errmsg | string | 错误信息 |

<br>

##### 2. dsl_get

接口: http://{{your_domain}}/dsl?project=%s

定义 DSL查询接口。接口根据project查询上面dsl_push上报的该模块所有dsl信息。若dsl_get为空，则 读取本机 [conf/dsl](../../../replayer-agent/conf/dsl) 下的配置文件。

| 接口说明 |  |
| :-----| :-----|
| Method | GET |
| Content-Type | application/x-www-form-urlencoded |

| 参数 | 类型 | 说明 |
| :-----| :----- | :----- |
| project | string | 必选，被测模块名 |

返回结果：dslData类型的数组

| dslData说明 |  |  |
| :-----| :-----| :-----|
| dsl | string | dsl信息，json串 |
| tag | string | 自定义名字 |
| project | string | 被测模块名 |
| addTime | string | 添加时间 |

<br>

##### 3. noise_push

接口: http://{{your_domain}}/noise

定义 噪音上报接口。在回放结果页，可以上报回放结果里的diff噪音。对于上报过的噪音，回放Agent会在对比回放结果时过滤对应噪音字段，提升回放成功率。若noise_push为空，则 存入本地 [conf/noise](../../../replayer-agent/conf/noise) 下的配置文件。

![push_noise](../../images/push_noise.png)

| 接口说明 |  |
| :-----| :-----|
| Method | POST |
| Content-Type | application/x-www-form-urlencoded |

| 参数 | 类型 | 说明 |
| :-----| :----- | :----- |
| project | string | 必选，被测模块名 |
| uri | string | 必选，噪音所属协议及标记，如http协议的uri |
| noise | string | 必选，要上报的噪音 |
| user | string | 可选，上报用户名 |

| 返回说明 |  |  |
| :-----| :-----| :-----|
| errno | int | 错误码，0 成功，非0 失败 |
| errmsg | string | 错误信息 |

<br>

##### 4. noise_del

接口: http://{{your_domain}}/noise/del

定义 噪音删除 接口。在回放详情页，删除曾经上报的噪音。若noise_del为空，则 删除本地 [conf/noise](../../../replayer-agent/conf/noise) 下配置文件内的噪音。
![del_noise](../../images/del_noise.png)

| 接口说明 |  |
| :-----| :-----|
| Method | POST |
| Content-Type | application/x-www-form-urlencoded |

| 参数 | 类型 | 说明 |
| :-----| :----- | :----- |
| id | string | 必选，噪音入库后的主键id |

| 返回说明 |  |  |
| :-----| :-----| :-----|
| errno | int | 错误码，0 成功，非0 失败 |
| errmsg | string | 错误信息 |

<br>

##### 5. noise_get

接口: http://{{your_domain}}/noise

定义 噪音查询接口。接口根据project和uri字段，查询上报的所有相关噪音数据。若noise_get为空，则 读取本地 [conf/noise](../../../replayer-agent/conf/noise) 下的配置文件。

| 接口说明 |  |
| :-----| :-----|
| Method | GET |
| Content-Type | application/x-www-form-urlencoded |

| 参数 | 类型 | 说明 |
| :-----| :----- | :----- |
| project | string | 必选，被测模块名 |
| uri | string | 必选，噪音所属协议及标记，如http协议的uri |

返回结果：NoiseInfo类型的数组

| NoiseInfo说明 |  |  |
| :-----| :-----| :-----|
| id | int | 存入数据库后的主键id |
| uri | string | 噪音所属协议及标记 |
| noise | string | 噪音 |
| project | string | 被测模块名 |
| addTime | string | 添加时间 |
| user | string | 上报用户 |

<br>

##### 6. module_info

接口: http://{{your_domain}}/platform/module?per=1000

定义 模块查询接口。即 查询已接入流量录制的模块信息。若module_info为空，则 读取本机配置文件 [conf/moduleinfo.json](../../../replayer-agent/conf/moduleinfo.json) 来获取模块信息。

| 接口说明 |  |
| :-----| :-----|
| Method | GET |
| Content-Type | application/x-www-form-urlencoded |

| 参数 | 类型 | 说明 |
| :-----| :----- | :----- |
| per | int | 返回最大结果数 |

返回结果：

{
    "data": []Module
}

| Module说明 |  |  |
| :-----| :-----| :-----|
| name | string | 接入模块名，即project |
| data | string | []KV类型数组的json串，存储模块详细信息 |

> 温馨提示：
>
> name尽量与编译后的$binName保持一致。如果存在冲突，可以增加前缀'\*-'，即\*-$binName。
>
> 如果name形如'\*a-b-c'，则尽量保证c具有可识别性，因为，[覆盖率统计回放](../replayer-codecov.md#1-覆盖率报告)会通过 *c* 来获取SUT进程信息并重启SUT。

| KV说明 |  |  |
| :-----| :-----| :-----|
| key | string | 与模块相关的一些信息，如监听地址listen-addr等 |
| value | string | 与key对应的具体值 |

目前KV类型中key支持的具体值有：
* context：必选，录制机器的地址，与录制流量的Context字段对应，以区分不同模块。一般一个模块找一台机器录制流量即可。

* listen-addr：必选，SUT的监听地址，一般为127.0.0.1:xxxx。

* department：可选，模块所属部门，默认空(则为default部门)。非空时，同时流量配置的读取自ES，则 会按部门字段读取es_url地址。es_url配置可详见：[回放Agent配置](../replayer-conf.md#5-es_url)

非本地回放时，Agent只用到上面三个key值，如果业务方想在该接口顺便存储模块其他信息，只需扩充[]KV即可。
