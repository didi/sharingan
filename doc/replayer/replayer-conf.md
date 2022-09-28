### Replayer-Agent 配置

<br>

Replayer-Agent的配置文件都在 **[./replayer-agent/conf](../../replayer-agent/conf)** 目录下。

* **conf/dsl目录**: 存放回放首页里 上报到本机的dsl数据，每个模块一个文件，以模块名为文件名。
* **conf/noise目录**: 存放回放结果页里 上报到本机的噪音数据，每个模块一个文件，以模块名为文件名。
* **conf/traffic目录**: 存放本机录制的测试流量，每个模块一个文件，以模块名为文件名。
* **conf/app.toml**: Replayer-Agent的核心配置文件，可以配置Web Server端口及超时时间，Mock Server端口、噪音及DSL上报到自有服务的http接口地址、流量查询的ES地址等。
* **conf/moduleinfo.json**: 模块配置。本地回放时，用来存放模块基本信息。

<br>

下面针对核心配置文件 conf/app.toml 的每个字段做详细的说明。

##### 1. [http]

定义 Web Server 的http相关配置信息。只需要注意端口字段，其他字段使用默认值即可。

* addr = ":8998"           # Replayer-Agent Web Server默认监听端口，可以修改
* handlerTimeout = 60000   # Handler timeout(ms), default 60000
* readHeaderTimeout = 2000 # Request header timeout(ms), default 2000
* readTimeout = 5000       # Receive http request timeout(ms), including the body, default 5000
* writeTimeout = 21000     # Receive http body and response timeout(ms), default 21000
* idleTimeout = 60000      # Keep-alive timeout(ms), default 60000

<br>

##### 2. [log]

定义 Web Server 的日志路径、文件名、保存时间

* filename      = "log/replayer.log.%Y%m%d%H" # 文件名格式
* linkname      = "log/replayer.log"          # 软连接
* maxHourAge    = 168                         # 默认日志保留168小时（7天）
* maxHourRotate = 1                           # 默认日志按小时切分

<br>

##### 3. [outbound]

定义 Mock Server 的默认地址，默认本机3515端口

* server_addr = "127.0.0.1:3515"

> 注意：
>
> 如需修改默认端口，则 需要在启动SUT时，同步定义REPLAYER_MOCK_PORT环境变量，否则回放会失败。

<br>

##### 4. [http_api]

非[本地回放](./README.md#4本地回放)时，用来定义 模块配置和上报噪音、DSL等信息的 自有服务接口，主要包括CRUD的http接口。接口设计详见[http接口说明](./conf/http_api.md)

* dsl_get = "http://{{your_domain}}/dsl?project=%s"
* dsl_push = "http://{{your_domain}}/dsl"
* noise_push = "http://{{your_domain}}/noise"
* noise_del = "http://{{your_domain}}/noise/del"
* noise_get = "http://{{your_domain}}/noise"
* module_info = "http://{{your_domain}}/platform/module?per=1000"

> 如果dsl_\*配置为空，则默认读取conf/dsl下的配置文件；
>
> 如果noise_\*配置为空，则默认读取conf/noise下的配置文件；
>
> 如果module_\*配置为空，则默认读取conf/mouleinfo.json；

<br>

##### 5. [es_url]

定义存储流量的es地址。

* default = "http://{{es_domain}}/xxx"

[本地回放](./README.md#4本地回放)： default 置空，会读取本机 [conf/traffic](../../replayer-agent/conf/traffic) 下的测试流量。

非本地回放： default 是 **必选** 字段，非空。


default字段非空时，支持不同业务部门的模块读取自不同的es地址。

> 如果 模块信息未配置department字段，即 未定义所属部门信息，则统一使用default配置的es地址。
> 
> 如果定义了所属部门信息，如department=Biz (参见 [回放模块配置](./conf/moduleinfo.md) example2。)，则新增下面配置即可。

* Biz = "http://{{es_domain}}/xxx"

> 注意:
>
> 使用Biz配置的前提是, default字段非空。所以，如果想使用部门字段地址，请确认default字段非空。

##### 6. [flow]

标识流量的一些基本信息。

* un_gzip = 0           # 回放的流量是否做了gzip解压；默认0未解压；1解压(mock服务会删除header Content-Encoding: gzip)
* line_max_size = 512   # 本地回放时，读取本地的每条流量最大512K, 不设置默认100K；如需更大，可以按需修改。

