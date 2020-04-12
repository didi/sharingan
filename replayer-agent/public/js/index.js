import * as menutab from "./menu.js"

var menu = new Vue({
    el: "#menu",
    data: function () {
        return {
            activeIndex: '1',
            userName: ""
        }
    },
    methods: {
        handleSelect: function(key, keyPath) {
            menutab.handleSelect(key, keyPath, this.activeIndex)
        },
        login: async function () {
            return true
        }
    },
});

var search = new Vue({
    el: "#search",
    data: function () {
        return {
            search: {
                project: 'passport',
                date: [new Date(), new Date()],
                page: 1,
                projects: [{
                    value: 'sps',
                    label: 'sps'
                }, {
                    value: 'passport',
                    label: 'passport'
                }]
            },
            picker_options: {
                shortcuts: [{
                    text: '今天',
                    onClick: function(picker) {
                        const end = new Date();
                        const start = new Date();
                        picker.$emit('pick', [start, end]);
                    }
                }, {
                    text: '昨天至今',
                    onClick: function(picker) {
                        const end = new Date();
                        const start = new Date();
                        start.setTime(start.getTime() - 3600 * 1000 * 24);
                        picker.$emit('pick', [start, end]);
                    }
                }, {
                    text: '一周前至今',
                    onClick: function(picker) {
                        const end = new Date();
                        const start = new Date();
                        start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
                        picker.$emit('pick', [start, end]);
                    }
                }, {
                    text: '30天前至今',
                    onClick: function(picker) {
                        const end = new Date();
                        const start = new Date();
                        start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
                        picker.$emit('pick', [start, end]);
                    }
                }],
                disabledDate: function(time) {
                    return time.getTime() > Date.now();
                }
            },

            ESDSL: "",
            search_show: true,
            dialogSubmitReport: false,
            dialogBatchReplay: false,

            reportFrom: {
                tag: ''
            },
            replayForm: {
                size: 2
            }
        }
    },
    methods: {
        changeProject: function () {
            this.setCookie("fd_project", this.search.project, 1000*60*60*24*30);
            this.loadDSL()
        },
        onSearch: function(page) {
            // 参数验证，目前只验证apollo格式是否正确
            if(!this.validate()) {
                return
            }
            flowList.loading = true
            flowList.currentPage = page
            this.search.page = page

            var that = this
            var tmpSearch = jQuery.extend(true, {}, this.search)
            delete tmpSearch["projects"]
            delete tmpSearch["heuristic"]

            axios({
                method: "post",
                url: "/search/",
                data: JSON.stringify(tmpSearch, true),
                headers: {'Content-Type': 'application/json'}
            }).then(function (response) {
                if(response.data.errmsg != "") {
                    that.$message.error(response.data.errmsg)
                }
                flowList.flowTable = response.data.results
                flowList.loading = false
            }).catch(function (error) {
                flowList.flowTable = []
                flowList.loading = false
                that.$message.error(error)
            });
        },
        validate: function() {
            if(this.search.apollo == undefined || this.search.apollo == "") {
                return true
            }
            var arr = this.search.apollo.split(" ")
            for(var i=0; i < arr.length; i++) {
                if(arr[i] == "1" || arr[i] == "0") {
                    this.$message.error("Apollo格式错误: 开关名[=状态]")
                    return false
                }
            }
            return true
        },
        submitReport: function (formName) {
            if(this.ESDSL == '{}' || this.reportFrom.tag == "") {
                this.$message.error("搜索或上报名称为空")
                return
            }
            var params = {}
            params.dsl = this.ESDSL
            params.project = this.search.project
            params.tag = jQuery.trim(this.reportFrom.tag)
            params.user = menu.userName
            var that = this
            axios({
                method: "post",
                url: "/platform/post/dsl",
                data: jQuery.param(params),
                headers: {'Content-Type': 'application/x-www-form-urlencoded'}
            }).then(function (response) {
                if(response.data.errno != 0) {
                    if(response.data.errmsg.startsWith("Error 1062: Duplicate")) {
                        that.$message.error("上报条件已存在,请勿重复上报!")
                        return
                    }
                    that.$message.error(response.data.errmsg)
                } else {
                    that.$message.success("上报成功")
                    that.dialogSubmitReport = false
                    that.loadDSL()
                }
            }).catch(function (error) {
                console.log(error)
                that.$message.error("上报失败")
            })
        },
        setESDSL: function(withDate) {
            var dsl = jQuery.extend(true, {}, this.search)
            delete dsl["projects"]
            if (withDate) {
                dsl.date = this.search.date;
            } else {
                dsl.date = undefined;
            }
            dsl.project = undefined
            dsl.page = undefined
            dsl.dsl = undefined
            dsl.heuristic = undefined
            jQuery.each(dsl, function(k, v){
                if(jQuery.trim(v) == "") {
                    dsl[k] = undefined
                }
            });
            this.ESDSL = JSON.stringify(dsl)
        },
        reportDSL: function (event) {
            this.setESDSL(false);
            this.dialogSubmitReport = true
        },
        batchReplay: function(event) {
            this.setESDSL(true);
            this.dialogBatchReplay = true
        },
        autoReplay: function(event) {
            var dsl = ""
            if(jQuery.trim(this.ESDSL).length > 3) {
                dsl = jQuery.trim(this.ESDSL)
            }
            window.open("/autoreplay/?project=" + this.search.project +"&dsl=" + encodeURIComponent(dsl) +"&size=" + jQuery.trim(this.replayForm.size), "_blank")
        },
        querySearch: function(queryString, cb) {
            var restaurants = this.restaurants;
            var results = queryString ? restaurants.filter(this.createFilter(queryString)) : restaurants;
            // 调用 callback 返回建议列表的数据
            cb(results);
        },
        createFilter: function(queryString) {
            return function(restaurant) {
                return (restaurant.value.toLowerCase().indexOf(queryString.toLowerCase()) === 0) ||
                    (restaurant.value.toLowerCase().indexOf("/"+queryString.toLowerCase()) !== -1);
            };
        },
        handleSelect: function (item) {
            this.search.inbound_request=''
            this.search.inbound_response=''
            this.search.outbound_request=''
            this.search.outbound_response=''
            this.search.apollo=''
            this.search.session_id=''
            delete this.search["heuristic"]
            jQuery.extend(this.search, jQuery.parseJSON(item.dsl))
            this.$forceUpdate()
        },
        loadModules: function() {
            var that = this
            axios({
                method: "get",
                url: "/platform/module/names",
            }).then(function (response) {
                that.search.projects = []
                for(var i=0; i < response.data.names.length; i++) {
                    that.search.projects[i] = []
                    that.search.projects[i]["value"] = response.data.names[i].name
                    that.search.projects[i]["label"] = response.data.names[i].name
                }
            }).catch(function (error) {
                console.log(error)
            })
        },
        loadDSL: function () {
            var params = {}
            params.project = this.search.project
            var res = []
            var that = this
            axios({
                method: "get",
                url: "/platform/get/dsl?" + jQuery.param(params),
                headers: {'Content-Type': 'application/x-www-form-urlencoded'}
            }).then(function (response) {
                that.restaurants = []
                for(var i=0; i < response.data.length; i++) {
                    that.restaurants[i] = []
                    that.restaurants[i]["value"] = response.data[i].tag
                    that.restaurants[i]["dsl"] = response.data[i].dsl
                }
            }).catch(function (error) {
                console.log(error)
            })
        },
        searchShow: function(event) {
            this.search_show = !this.search_show
        },
    },
    mounted: function() {
        var pro = this.getCookie("fd_project")
        if (pro) {
            this.search.project = pro
        }
        var now = new Date()
        var m = now.getMonth() + 1
        if(m < 10) {
            m = "0" + m
        }
        var d = now.getDate()
        if(d < 10) {
            d = "0" + d
        }
        var today = now.getFullYear()+"-" + m + "-" + d;
        this.search.date = [today, today];
        this.loadDSL()
        this.loadModules()

        var that = this
        // menu.login().then(function (response) {
        //     if (response.data.user=='') {
        //         setTimeout(function(){document.location.href = response.data.location;},100);
        //         return;
        //     }
        //     menu.userName = response.data.user;
        //     // search on startup
        //     that.onSearch(1)
        // }).catch(function (error) {
        //     console.log(error)
        // });
        menu.login().then(function () {
            menu.userName = "";
            // search on startup
            that.onSearch(1)
        }).catch(function (error) {
            console.log(error)
        });
    }
});

var flowList = new Vue({
    el: "#flow_list",
    data: function() {
        return {
            flowTable: [],
            loading: false,
            currentPage: 1
        }
    },
    methods: {
        getActions: function (actions) {
            return actions
        },
        masterReq: function (row) {
            if (row.rowIndex === 0) {
                return "primary-row"
            }
            return ''
        },
        jump: function (row) {
            window.open("/replay/" + row.sessionId + "?project="+row.project, "_blank")
        },
        handleCurrentChange: function (page) {
            search.onSearch(page)
        }
    }
});

