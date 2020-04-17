### 本地回放接入

<br>

Replayer-Agent默认的接入方式是[本地回放](./README.md#4本地回放)。如需 非本地回放，只需根据 [Replayer-Agent配置](./replayer-conf.md#4-http_api) 设置[http_api]和[es_url]接口地址即可。

<br>

对于有 新增 本地回放模块的需求同学，可以根据如下3步操作即可。

>步骤a. 增加模块信息

在 [conf/moduleinfo.json](../../replayer-agent/conf/moduleinfo.json) 内增加模块基本信息。字段参考 [回放模块配置](./conf/moduleinfo.md)

<br>

>步骤b. 修改app.toml

修改 [conf/app.toml](../../replayer-agent/conf/app.toml) 配置，让流量、DSL、噪音等优先读写本地配置文件。字段详见: [Replayer-Agent配置](./replayer-conf.md)

  * 注释掉http_api下的所有字段 (噪音/DSL读取本地)
  * 注释掉es_url下的所有字段 (流量读取本地)

<br>

>步骤c. 增加录制流量

将录制流量存入 [conf/traffic](../../replayer-agent/conf/traffic) 下，文件名为模块project值

<br>

**至此，就可以开始 仅依赖本地配置文件的 本地回放啦~**

<br>

> 温馨提示：
>
>模块、噪音、DSL、流量可以独立配置。比如 只配置流量读取es，其他都读取本地文件。