import * as mutex from "./mutex.js"

const TypeInit = 1;
const TypeReplay = 2;
const TimeSleep = 0

;(function() {
    var auto = new Vue({
        el: "#batchReplay",
        data: function() {
            return {
                timer: null,
                replays: [],
                badcase: {
                    data: [{
                        label: 'Bad Case',
                        children: []
                    }],
                    props: {
                        children: 'children',
                        label: 'label'
                    }
                },
                loading: true,
                ops: {
                    retry: {},      // 记录重试次数
                    history: {},    // 记录已回放结果，避免重复
                    size: 0,        // 每个dsl回放case数
                    date: [],       // 记录时间区间
                    parallel: 0,    // 并发计数
                    dsls: [],       // 回放dsl集合
                    dslIdx: 0,      // 当前回放dsl游标
                    sessions:[],    // 记录当前dsl的case集合
                    sessionIdx: 0,  // 当前dsl回放case游标
                },
                stats: {
                    estcnt: 0,  // 估算回放case总数
                    dslcnt: 0,  // 回放dsl总数
                    casecnt: 0, // 回放case总数
                    failcnt: 0, // 回放case失败数
                },
            }
        },
        methods: {
            getPercentage: function() {
                if(this.stats.estcnt == 0) {
                    return 0
                }
                return (100*this.stats.casecnt/this.stats.estcnt).toFixed(2)
            },
            replayParallel: function() {
                this.ops.size = parseInt(Global.Size, 10)
                if (Global.Dsl == "") {
                    var that = this
                    axios({
                        method: "get",
                        url: "/platform/get/dsl?" + jQuery.param({project: Global.Project}),
                        headers: {'Content-Type': 'application/x-www-form-urlencoded'}
                    }).then(function (response) {
                        that.ops.dsls = that.filterRecommend(response.data)
                        that.stats.estcnt = that.ops.dsls.length * that.ops.size
                        that.ops.mutex = new mutex.Mutex()
                        that.replayNext(TypeInit)
                    }).catch(function (error) {
                        console.error(error)
                        that.showDialog()
                    })
                } else {
                    var dslObj = JSON.parse(Global.Dsl);
                    this.ops.date = dslObj.date

                    this.ops.dsls.push({dsl: Global.Dsl})
                    this.stats.estcnt = this.ops.size
                    this.ops.mutex = new mutex.Mutex()
                    this.replayNext(TypeInit)
                }
            },
            replayNext: function(typ) {
                var that = this
                this.ops.mutex.lock(function () {
                    if(typ == TypeInit || that.ops.sessionIdx >= that.ops.sessions.length) {
                        // exit entry
                        if(that.ops.dslIdx >= that.ops.dsls.length) {
                            //that.showDialog()
                            that.ops.parallel--
                            if(that.ops.parallel <= 0) {
                                that.showDialog()
                            }
                            that.ops.mutex.unlock()
                            return
                        }

                        var dsl = that.ops.dsls[that.ops.dslIdx].dsl
                        // search for the next dsl
                        that.search(dsl).then(function(result) {
                            if (that.ops.size > result.data.results.length) {
                                that.stats.estcnt = that.stats.estcnt - (that.ops.size - result.data.results.length)
                                that.ops.sessions = result.data.results
                            } else {
                                that.ops.sessions = result.data.results.slice(0, that.ops.size)
                            }
                            that.ops.dslIdx++
                            that.ops.sessionIdx = 0

                            if(typ == TypeInit) {
                                if (that.ops.sessions.length < result.data.parallel) {
                                    that.ops.parallel = that.ops.sessions.length
                                } else {
                                    that.ops.parallel = result.data.parallel
                                }
                                that.ops.mutex.unlock()
                                //that.replayNext(TypeReplay)
                                for(var i = 0; i < that.ops.parallel; i++) {
                                    that.replayNext(TypeReplay)
                                }
                                return
                            }

                            if(result.data.results.length == 0) {
                                that.replays.push({dsl: dsl, result: 1})
                                that.loading = false
                                that.ops.mutex.unlock()
                                //that.replayNext(typ)
                                that.timer = setTimeout(()=>{   //设置延迟执行
                                    that.replayNext(typ)
                                }, TimeSleep);
                                return
                            }
                            that.replay(dsl, that.ops.sessions[that.ops.sessionIdx].sessionId)

                        }, function(error) {
                            console.error(error)
                            that.stats.estcnt = that.stats.estcnt - that.ops.size
                            that.replays.push({dsl: dsl, result: 1})
                            that.ops.dslIdx++
                            that.ops.mutex.unlock()
                            //that.replayNext(typ)
                            that.timer = setTimeout(()=>{   //设置延迟执行
                                that.replayNext(typ)
                            }, TimeSleep);
                        })
                    } else {
                        that.replay(that.ops.dsls[that.ops.dslIdx - 1].dsl, that.ops.sessions[that.ops.sessionIdx].sessionId)
                    }
                });
            },
            replay: function(dsl, sid) {
                //this.ops.sessionIdx++
                if (this.ops.retry[sid] == undefined) {
                    this.ops.retry[sid] = 1
                    this.ops.sessionIdx++
                }
                //if(this.ops.history[sid] != undefined) {
                if(this.ops.history[sid] != undefined && this.ops.retry[sid] == 1) {
                    this.ops.sessionIdx++
                    this.replays.push({dsl: dsl + " (duplicated: " + sid + " )", result: 1})
                    this.stats.estcnt = this.stats.estcnt - 1
                    this.ops.mutex.unlock()
                    //this.replayNext(TypeReplay)
                    var that = this
                    this.timer = setTimeout(()=>{   //设置延迟执行
                        that.replayNext(TypeReplay)
                    }, TimeSleep);
                    return
                }
                this.ops.history[sid] = true
                this.ops.mutex.unlock()

                var that = this
                axios({
                    method: "GET",
                    url: "/replayed/"+sid,
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    params: {
                        retType: "witherror",
                        project: Global.Project,
                    },
                }).then(function(response) {
                    var result = 0
                    if(!response.data.success) {
                        if (that.ops.retry[sid]<=3) {
                            that.ops.retry[sid] = that.ops.retry[sid] + 1
                            that.ops.mutex.lock(function () {
                                that.timer = setTimeout(()=>{   //设置延迟执行
                                    console.log("sid=", sid, " retry=", that.ops.retry[sid])
                                    that.replay(dsl, sid)
                                }, TimeSleep);
                            })
                            return
                        }
                        result = 2
                        that.stats.failcnt++
                        that.badcase.data[0].children.push({label: sid})
                    }
                    that.stats.casecnt++
                    that.replays.push({dsl: dsl, project: Global.Project, sessionId: sid, result: result})
                    that.loading = false
                    //that.replayNext(TypeReplay)
                    that.timer = setTimeout(()=>{   //设置延迟执行
                        that.replayNext(TypeReplay)
                    }, TimeSleep);
                }).catch(function (error) {
                    console.error(error)
                    that.stats.failcnt++
                    that.badcase.data[0].children.push({label: sid})
                    that.stats.casecnt++
                    that.replays.push({dsl: dsl, project: Global.Project, sessionId: sid, result: 2})
                    that.loading = false
                    //that.replayNext(TypeReplay)
                    that.timer = setTimeout(()=>{   //设置延迟执行
                        that.replayNext(TypeReplay)
                    }, TimeSleep);
                })
            },
            search: async function(dsl) {
                try {
                    var params = JSON.parse(dsl)
                    params.project = Global.Project
                    params.page = 1
                    params.size = this.ops.size
                    if (params.date == "" && this.ops.date != "" ) {
                        params.date = this.ops.date
                    }
                    params.field = ["SessionId"]
                    return await axios({
                        method: "post",
                        url: "/search/",
                        data: JSON.stringify(params, true),
                        headers: {'Content-Type': 'application/json'}
                    })
                } catch(error) {
                    console.error(error)
                }
                return {}
            },
            jump: function(node) {
                setTimeout(function(){window.open("/replay/" + node.data.label + "?project=" + Global.Project, "_blank");}, 100)
            },
            filterRecommend: function(dsls) {
                var recs = []
                for(var i=0; i < dsls.length; i++) {
                    if (dsls[i].recommend == 1 || typeof(dsls[i].recommend) == "undefined") {
                        recs.push(dsls[i])
                    }
                }
                return recs
            },
            showDialog: function() {
                this.loading = false
                var msg = '回放完成: {回放总数:' + this.stats.casecnt + ', 其中失败数:' + this.stats.failcnt + '}'
                if(this.stats.casecnt == 0) {
                    msg = '数据为空'
                }
                this.$notify({
                    title: '提示',
                    message: msg+'</br><a href="/coverage?project='+ Global.Project +'" target="_blank"><button type="button" class="el-button el-button--text"><span>覆盖率报告</span></button></a>&nbsp;&nbsp;<a href="https://github.com/didi/sharingan/blob/master/doc/replayer/replayer-codecov.md" target="_blank"><button type="button" class="el-button el-button--text"><span>覆盖率使用手册</span></button></a>',
                    type: 'success',
                    dangerouslyUseHTMLString: true,
                    offset: 100,
                    duration: 0
                });
                return
            }
        },
        mounted: function() {
            this.replayParallel()
        }
    })
})()
