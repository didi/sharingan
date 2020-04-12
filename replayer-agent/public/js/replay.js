;(function() {
    var session = new Vue({
        el: "#session",
        data: function() {
            return {
                sessionId: "",
                actions: [],/*[{
                    diff: diffHtml,
                    onlineReq: "online request",
                    testReq: "testing request",
                    onlineRes: "online response",
                    testRes: "testing response"
                }]*/
                // 整体回放结果
                result: 1,
                data5: [{is:0, "label":"一级", children:[{"name":"w",label:"二级"},{"name":"w",label:"二级"}]},
                        {is:1, "label":"一级", children:[{"name":"w",label:"二级"},{"name":"w",label:"二级"}]},
                        {is:2, "label":"一级", children:[{"name":"w",label:"二级"},{"name":"w",label:"二级"}]}
                    ],
                userName: '',
                expands:[],
                dialogVisible: false,
            }
        },
        methods: {
            cellStyle: function( row, column, rowIndex, columnIndex ) {
                if (row.row.scorePercentage == "主请求") {
                    return 'font-weight:bold;font-size:13px;'
                } else if (row.columnIndex == 0) {
                    return 'padding-left: 2%'
                } else if (row.columnIndex == 1) {
                    return 'padding-left: 1%'
                } else {
                    return ''
                }
             },
            // getRowClass({ row, column, rowIndex, columnIndex }) {
            //     if (rowIndex == 0) {
            //         return 'background:rgba(64,158,255,.1)'
            //     } else {
            //         return ''
            //     }
            // },
            filterHandler(value, row, column) {
                const property = column['property'];
                return row[property] === value;
            },
            clearFilter() {
                this.$refs.filterTable.clearFilter();
            },
            rowColor: function(row) {
                if(row.row.isDiff == 1) {
                    return "warning-row"
                } else if(row.row.isDiff == 2) {
                    return "error-row"
                }
                return 'success-row'
            },
            reportNoise: function(project, uri, noise, node) {
                var param = {
                    project: project,
                    uri, uri,
                    noise: noise,
                    user: this.userName
                }
                axios({
                    method: "post",
                    url: "/noise/",
                    data: jQuery.param(param),
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'}
                }).then(function (response) {
                    node.remove()
                    console.log(response)
                }).catch(function (error) {
                    console.log(error)
                })
            },
            delNoise: function(node) {
                if (node.data.noiseId == 0 && node.data.noiseUri == "" && node.data.noiseData == "" && node.data.noiseProject == "") {
                    // ignored by global rules
                    this.$message.warning("ignored by code, cannot be cancelled!")
                    return
                }
                var param = {
                    id: node.data.noiseId,
                    project: node.data.noiseProject,
                    uri: node.data.noiseUri,
                    noise: node.data.noiseData
                }
                axios({
                    method: "post",
                    url: "/noise/del",
                    data: jQuery.param(param),
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'}
                }).then(function (response) {
                    node.remove()
                    console.log(response)
                }).catch(function (error) {
                    console.log(error)
                })
            },
            diffBinary: function(isBinary, action, onlineBytes, testBytes) {
                if(isBinary || action.row.diff != '') {
                    return
                }

                var params = {
                    online: onlineBytes,
                    test: testBytes,
                }
                axios({
                    method: "post",
                    url: "/diffbinary",
                    data: JSON.stringify(params, true),
                    headers: {'Content-Type': 'application/json'}
                }).then(function (response) {
                    action.row.diff = Diff2Html.getPrettySideBySideHtmlFromDiff(response.data.diff)
                }).catch(function (error) {
                    console.log(error)
                })
            },
            isMocked: function(row) {
                return row.mockedRes.length != 0
            },
            base64: function(isBinary, action, key, bytes) {
                if(isBinary) {
                    var param = {
                        base64: bytes
                    }
                    axios({
                        method: "post",
                        url: "/xxd",
                        data: jQuery.param(param),
                        headers: {'Content-Type': 'application/x-www-form-urlencoded'}
                    }).then(function (response) {
                        action.row[key] = response.data
                    }).catch(function (error) {
                        console.log(error)
                    })
                }
            },
            getRowKeys: function(row) {
                return row.id+""
            },
            responseFormatter: function(row, column) {
                if(row.mockedRes != "") {
                    return row.mockedRes
                }
                return row.onlineRes
            },
            // traceStack: function() {
            //     var that = this
            //     const loading = this.$loading({})
            //     axios({
            //         method: "get",
            //         url: "/replayed/"+Global.sid,
            //         params: {
            //             trace:"on",
            //             project:Global.Project,
            //         },
            //         headers: {'Content-Type': 'application/x-www-form-urlencoded'}
            //     }).then(function (response) {
            //         var diffs = response.data.diffs
            //         for (var i=0; i<diffs.length; i++) {
            //             diffs[i].diffBinary = true
            //             diffs[i].onlineReqBinary = false
            //             diffs[i].onlineResBinary = false
            //             diffs[i].testReqBinary = false
            //             diffs[i].testResBinary = false
            //         }
            //         if(response.data.success) {
            //             that.result = 0
            //         } else {
            //             that.result = 2
            //         }
            //         that.actions = diffs
            //         loading.close()
            //         window.open("/trace?type=stack", "_blank")
            //     }).catch(function (error) {
            //         console.log(error)
            //         loading.close()
            //     })
            // },
            onmouseup: function() {
                // 获取选中的内容
                var selectData = ""
                if(document.selection){
                    selectData = document.selection.createRange().text.toString().trim(); // ie浏览器
                } else {
                    selectData = document.getSelection().toString().trim();
                }

                if(!selectData){ return ; }

                // 只显示json格式
                try{ JSON.parse(selectData); } catch(err){ return ; }

                this.dialogVisible = true

                // 延时执行
                setTimeout(function() {
                    $(".el-dialog__body").css({
                        "max-height":"400px",
                        "overflow":"auto",
                        "padding":"0px 10px"
                    }).jsonFormat({
                        expanded:"/public/image/expanded.gif",
                        collapsed:"/public/image/collapsed.gif",
                        data:selectData
                    });
                }, 0)
            }
        },
        mounted: function() {
            this.sessionId = Global.sid
            var that = this
            const loading = this.$loading({})
            function replay() {
                return axios({
                    method: "get",
                    url: "/replayed/"+Global.sid,
                    params: {
                        project:Global.Project,
                    },
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'}
                })
            }

            function login() {
                return true
            }

            axios.all([replay(), login()]).then(axios.spread(function (response, userInfo) {
                // record user info
                // if (userInfo.data.user=='') {
                //     setTimeout(function(){document.location.href = userInfo.data.location;},100);
                //     return;
                // }
                //that.userName = userInfo.data.user;
                that.userName = "";

                if(response.data.success) {
                    that.result = 0
                } else {
                    that.result = 2
                }
                if(response.data.errmsg != "") {
                    console.error(response.data.errmsg)
                    that.$message.error(response.data.errmsg)
                }

                var project = '';
                var diffs = response.data.diffs;
                var diffsRes = [];
                // fill the page with replayed session
                for (var i=0; diffs != undefined && i<diffs.length; i++) {
                    if (diffs[i].noWebDisplay == true) {
                        continue
                    }
                    diffs[i].diffBinary = true
                    diffs[i].onlineReqBinary = false
                    diffs[i].onlineResBinary = false
                    diffs[i].testReqBinary = false
                    diffs[i].testResBinary = false
                    if (project == '') {
                        project = diffs[i].project;
                    }
                    diffsRes.push(diffs[i])
                }
                diffsRes[0].scorePercentage = "主请求"
                that.actions = diffsRes
                loading.close()

                // if (project != "" && that.userName != "") {
                //     metric.Up(metric.AppFD2, [metric.ActReplay], {
                //         "project": project,
                //         "user": that.userName,
                //         "version": Global.Version,
                //     })
                // }
                that.expands.push("0")
            })).catch(function (error) {
                loading.close()
                console.log(error)
            })
        }
    })
})();
