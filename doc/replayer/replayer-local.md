### 本地回放接入

回放Agent默认的接入方式是[本地回放](./README.md#4本地回放)。如需 非本地回放，只需根据 [回放Agent配置](./replayer-conf.md#4-[http_api]) 设置[http_api]和[es_url]接口地址即可。

对于有 新增 本地回放模块的同学，可以根据如下接入流程操作即可。
##### 1. 接入流程

>步骤a. 增加模块信息

在 [conf/moduleinfo.json](../../replayer-agent/conf/moduleinfo.json) 内增加模块基本信息。字段参考 [回放模块配置](./conf/moduleinfo.md)

>步骤b. 修改app.toml

修改 [conf/app.toml](../../replayer-agent/conf/app.toml) 配置，让流量、DSL、噪音等优先读写本地配置文件，而不是http接口。

1. 注释掉http_api下的所有字段
2. 注释掉es_url下的所有字段

各配置字段含义详见: [回放Agent配置](./replayer-conf.md)

>步骤c. 增加录制流量

将录制流量存入 [conf/traffic](../../replayer-agent/conf/traffic) 下，文件名为模块project值

**至此，就可以开始 仅依赖本地配置文件的 本地回放啦~**
