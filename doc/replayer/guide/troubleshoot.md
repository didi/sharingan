### 常见问题及排查

<br>

sharingan已经在滴滴内部服务了一批go业务模块，现将各业务模块接入和回放过程遇到的问题整理如下，方便大家及时解决问题。

<br>

具体问题如下：

##### 1. 回放页面Fail且无流量
![prob_sut_notstart](http://img-hxy021.didistatic.com/static/sharingan/prob_sut_notstart_v2.png)

* **现象**：如上图，回放页面整体飘红，且顶端提示 "dial tcp xxx connect: connection refused"。

* **原因**：SUT未启动

* **解决**：请启动SUT服务

<br>

##### 2. SUT启动失败
* **现象1**：SUT启动时，提示TCP连接:3515拒绝

* **原因**：Replayer-Agent服务未启动

* **解决**：请先启动Replayer-Agent，再启动SUT服务

***

* **现象2**：SUT启动时，提示连接池配置相关的TCP连接xx超时；一般为mysql、redis等连接池配置

* **原因**：SUT启动阶段，Mock Server对任何TCP连接只起代理作用

* **解决**：请修改SUT连接池配置，为本机可以连通的有效ip:port

<br>

##### 3. 所有Outbound回放失败，'Testing Request'提示missing
* **现象**：除Inbound外，所有Outbound都飘红，且点开流量'Testing Request'提示missing。

* **原因**：SUT未import replayer包，且未使用定制golang编译代码

* **解决**：请按 [回放接入-2. 配置并启动SUT](../README.md#2-配置并启动sut) 操作。

<br>

##### 4. 部分Outbound回放失败，'O==T Diff'提示存在diff
* **现象**：部分Outbound流量回放失败，且点开流量'Testing Request'未提示missing，而是存在diff

* **可能原因1**：回放的流量时间日期太旧，被测代码逻辑跟录制流量时的代码逻辑已经差别很大

* **解决**：请重新录制最新代码流量

<br>

* **可能原因2**：diff属于噪音

* **解决**：请上报噪音，重新回放

<br>

* **可能原因3**：被测代码逻辑有修改

* **解决**：确认符合修改预期即可

<br>

* **可能原因4**：流量本身是异常的；如 录制流量的机器，有人手动调试代码

* **解决**：尝试换个session，如果回放正常，请忽略此异常流量

<br>

##### 5. 部分Outbound匹配失败，'O==T Diff'没有diff
****首先 确认回放的流量是否太旧!!!****

* **现象1**：'Testing Request'提示missing

* **可能原因1**：线下配置导致线下少发部分请求，即回放流量里多出来一些Outbound子请求没有被匹配。

* **解决**：请保持线下配置与线上配置一致

<br>

* **可能原因2**：被测代码逻辑有修改，减少了Outbound请求。

* **解决**：确认符合修改预期即可

***

* **现象2**：'Online Request'提示not matched

* **可能原因1**：线下配置导致线下多发部分请求，即线下多出来一些Outbound子请求没有被匹配。

* **解决**：请保持线下配置与线上配置一致

<br>

* **可能原因2**：被测代码逻辑有修改，增加了Outbound请求。

* **解决**：确认符合修改预期即可

<br>

##### 6. 点击"覆盖率报告"链接
* **现象1**：提示*.test未启动

* **原因**：SUT启动方式为普通回放

* **解决**：请参考 [覆盖率统计回放](../replayer-codecov.md) 重启SUT

*** 

* **现象2**：覆盖率结果很低

* **可能原因1**：点击"覆盖率报告"前，未回放任何流量。

* **解决**：请回放一个或多个流量后，再点击"覆盖率报告"链接

<br>

* **可能原因2**：连续点击"覆盖率报告"链接。每次点击"覆盖率报告"链接，Replayer-Agent会自动重启SUT服务，重新统计覆盖率。

* **解决**：请在两次点击"覆盖率报告"链接之间，回放一个或多个流量。[覆盖率统计回放-1. 覆盖率报告](../replayer-codecov.md#1-覆盖率报告)

<br>

* **可能原因3**：SUT编译时，设置的统计整个模块代码，但回放的流量只涉及部分目录。

* **解决**：参考 [覆盖率统计回放-2. 配置并启动SUT](../replayer-codecov.md#2-配置并启动SUT) 修改编译命令，设置-coverpkg为指定目录 

<br>
<br>

##### 交流群

除了上面的问题，欢迎大家加入 **Sharingan QQ交流群**，一起交流~
<br>

![QQ](http://img-hxy021.didistatic.com/static/sharingan/QQ_v2.JPG)
