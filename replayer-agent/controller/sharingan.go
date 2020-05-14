package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/global"
	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/common/handlers/httpclient"
	"github.com/didi/sharingan/replayer-agent/common/handlers/module"
	"github.com/didi/sharingan/replayer-agent/common/handlers/outbound"
	"github.com/didi/sharingan/replayer-agent/common/handlers/template"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/idl"
	"github.com/didi/sharingan/replayer-agent/logic/bind"
	"github.com/didi/sharingan/replayer-agent/logic/replayed"
	"github.com/didi/sharingan/replayer-agent/logic/search"
	"github.com/didi/sharingan/replayer-agent/logic/worker"
	"github.com/didi/sharingan/replayer-agent/model/esmodel"
	"github.com/didi/sharingan/replayer-agent/model/nuwaplt"
	"github.com/didi/sharingan/replayer-agent/utils/helper"

	jsoniter "github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
)

type ShaRinGan struct {
	BaseController
	mu sync.Mutex
}

/**
 * 首页
 */
func (srg ShaRinGan) Index(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	// 更新模块信息，刷新本地缓存
	nuwaplt.Reload()

	data := struct{ Version string }{global.Version}
	template.Render(w, "index", data)
}

/**
 * 流量搜索
 */
func (srg ShaRinGan) Search(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	var req idl.SearchReq
	if err := bind.Bind(r, &req); err != nil {
		srg.EchoJSON(w, r, idl.SearchResp{Errmsg: "parse params failed"})
		return
	}

	ctx := r.Context()

	parallel := global.FlagHandler.Parallel
	_, ok := nuwaplt.Module2Host[req.Project]
	if ok {
		// get department info
		depart := nuwaplt.GetValueByKey(req.Project, nuwaplt.KDepartment, nuwaplt.DefaultDepartment)
		ctx = context.WithValue(ctx, nuwaplt.KDepartment, depart)
	}

	data := search.Search(ctx, &req)
	srg.EchoJSON(w, r, idl.SearchResp{Results: data, Parallel: parallel})
}

/**
 * 回放平台
 */
func (srg ShaRinGan) Replay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conf.FreshHandler()
	sid := ps.ByName("sessionId")
	r.ParseForm()
	project := r.Form.Get("project")

	data := struct{ Sid, Version, Project string }{sid, global.Version, project}
	template.Render(w, "replay", data)
}

/**
 * 单个session回放
 *
 * @Return ajax返回
 */
func (srg ShaRinGan) Replayed(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conf.FreshHandler()
	var req idl.ReplayedReq
	if err := bind.Bind(r, &req); err != nil {
		srg.EchoJSON(w, r, idl.ReplayedResp{Success: false, Errmsg: "parse params failed"})
		return
	}

	// set department information for choosing the corresponding es
	sid := ps.ByName("sessionId")
	depart := nuwaplt.GetValueWithProject(req.Project, nuwaplt.KDepartment, nuwaplt.DefaultDepartment)
	ctx := context.WithValue(r.Context(), nuwaplt.KDepartment, depart)

	// fetch sessions
	sessions := worker.FetchSessions(ctx, sid, req.Project)
	if len(sessions) <= 0 {
		srg.EchoJSON(w, r, idl.ReplayedResp{Success: false, Errmsg: "search session failed"})
		return
	}
	session := sessions[0]

	// listenAddr
	listenAddr := nuwaplt.GetValueByKey(req.Project, nuwaplt.KListenAddr, global.ListenAddr)
	if listenAddr == "" {
		srg.EchoJSON(w, r, idl.ReplayedResp{Success: false, Errmsg: "get listen addr of SUT failed"})
		return
	}

	// to replay
	replayer := &worker.Replayer{BasePort: outbound.BasePort}
	replayer.Protocol = nuwaplt.GetValueByKey(req.Project, nuwaplt.KProtocol, nuwaplt.PHttp)
	replayer.Language = nuwaplt.GetValueByKey(req.Project, nuwaplt.KLanguage, nuwaplt.LGO)
	replayer.ReplayAddr = listenAddr

	// begin replay【ShaRinGan -> sut -> mock server】
	err := replayer.ReplaySession(ctx, session, req.Project)
	if err != nil {
		srg.EchoJSON(w, r, idl.ReplayedResp{Success: false, Errmsg: err.Error()})
		return
	}

	diffs := replayed.DiffReplayed(ctx, replayer.ReplayedSession, req.Project)

	resp := idl.ReplayedResp{Success: replayed.Judge(diffs)}
	switch req.RetType {
	case "simple":
		// return no diffs
	case "witherror":
		// return diffs with error type
		resp.Diffs = replayed.QDEFormat(sid, resp.Success, diffs)
	default:
		// return all the diffs
		resp.Diffs = diffs
	}
	srg.EchoJSON(w, r, resp)
}

/**
 * 批量回放
 */
func (srg ShaRinGan) AutoReplay(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	r.ParseForm()
	project := r.Form.Get("project")
	dsl := r.Form.Get("dsl")
	size := r.Form.Get("size")

	data := struct{ Project, Dsl, Size, Version string }{project, dsl, size, global.Version}
	template.Render(w, "auto", data)
}

/**
 * 噪音上报
 */
func (srg ShaRinGan) Noise(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	project := r.PostFormValue("project")
	uri := r.PostFormValue("uri")
	noise := r.PostFormValue("noise")
	user := r.PostFormValue("user")

	result := nuwaplt.ReportNoise(project, uri, noise, user)
	srg.EchoJSON(w, r, result)
}

/**
 * 删除噪音
 */
func (srg ShaRinGan) DelNoise(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	id := r.PostFormValue("id")
	project := r.PostFormValue("project")
	uri := r.PostFormValue("uri")
	noise := r.PostFormValue("noise")

	result := nuwaplt.DelNoise(id, project, uri, noise)
	srg.EchoJSON(w, r, result)
}

/**
 * 查看session详情
 */
func (srg ShaRinGan) Session(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conf.FreshHandler()
	r.ParseForm()
	project := r.Form.Get("project")
	depart := nuwaplt.GetValueWithProject(project, nuwaplt.KDepartment, nuwaplt.DefaultDepartment)
	ctx := context.WithValue(r.Context(), nuwaplt.KDepartment, depart)
	sid := ps.ByName("sessionId")

	var bodyByte []byte
	var err error
	//优先读取es地址
	if conf.Handler.GetString("es_url.default") != "" {
		cond := &idl.SearchReq{SessionId: sid, Size: 1}
		bodyByte, err = search.Query(ctx, cond, 0)
		if err != nil {
			srg.Echo(w, r, bodyByte)
			return
		}
	} else {
		//读取配置文件conf/traffic/{project}
		contents, err := helper.ReadLines(conf.Root + "/conf/traffic/" + project)
		if err != nil {
			tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, "Failed to read /conf/traffic/"+project+", err="+err.Error())
			srg.Echo(w, r, bodyByte)
			return
		}
		var json = jsoniter.ConfigCompatibleWithStandardLibrary

		//原始流量格式
		for _, flow := range contents {
			traffic := &esmodel.SessionId{}
			err := json.Unmarshal([]byte(flow), traffic)
			if err != nil {
				tlog.Handler.Warnf(r.Context(), tlog.DLTagUndefined, "errmsg= Failed at UmMarshal origin traffic||err=%s", err.Error())
				continue
			}
			if traffic.Id == sid {
				bodyByte = []byte(flow)
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(bodyByte) != 0 {
		srg.Echo(w, r, bodyByte)
	} else {
		srg.Echo(w, r, []byte("No Result, please change another sessionId~"))
	}

}

// base64 decode and bianry
func (srg ShaRinGan) Xxd(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	binaryData := r.PostFormValue("base64")
	decodeBinary, decodeErr := base64.StdEncoding.DecodeString(binaryData)
	if decodeErr != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "errmsg=base64 decode failed||err=%s", decodeErr)
	}

	result, binaryErr := replayed.ToBinary(decodeBinary)
	if binaryErr != nil {
		tlog.Handler.Warnf(r.Context(), tlog.DLTagUndefined, "errmsg=binary fail||err=%s", binaryErr)
	}
	srg.Echo(w, r, result)
}

func (srg ShaRinGan) DiffBinary(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	var request *idl.DiffBinaryReq
	var response *idl.DiffBinaryResp

	request = &idl.DiffBinaryReq{}
	err := bind.Bind(r, request)
	if err != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "binding request to struct failed: %s", err)
		return
	}
	diff, err := replayed.BinaryDiff(helper.StringToBytes(request.Online), helper.StringToBytes(request.Test))
	response = &idl.DiffBinaryResp{Diff: helper.BytesToString(diff)}

	srg.EchoJSON(w, r, response)
}

type modules struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
	Total  int    `json:"total"`
	//Data []string `json:"data"`
	Names []nuwaplt.Name `json:"names"`
}

//PlatformModules nuwa平台接口 读取所有模块数据
func (srg ShaRinGan) PlatformModules(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	res := modules{}
	//优先读取http接口 conf/app.json下的http_api.module_info  再读取配置文件conf/moduleinfo.json
	nuwaplt.Reload()
	res.Names = nuwaplt.ModuleNames
	res.Total = len(res.Names)

	srg.EchoJSON(w, r, res)
}

type dslData struct {
	Dsl     string `json:"dsl"`
	Tag     string `json:"tag"`
	Project string `json:"project"`
	AddTime string `json:"addTime"`
}

//PlatformGetDsl nuwa平台接口 查询模块dsl上报数据
func (srg ShaRinGan) PlatformGetDsl(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	resDefault := make([]dslData, 0)
	res := make([]dslData, 0)
	project := r.FormValue("project")
	if project == "" {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Params project is illegal!")
		srg.EchoJSON(w, r, resDefault)
		return
	}
	//优先读取http接口
	if url := conf.Handler.GetString("http_api.dsl_get"); url != "" {
		_, bodyBytes, err := httpclient.Handler.Get(r.Context(), fmt.Sprintf(url, project))
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to curl "+url+", "+err.Error())
			srg.EchoJSON(w, r, resDefault)
			return
		}
		err = json.Unmarshal(bodyBytes, &res)
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to unmarshal "+url+", "+err.Error())
			srg.EchoJSON(w, r, resDefault)
			return
		}

	} else {
		//读取配置文件conf/dsl/{project}
		contents, err := helper.ReadLines(conf.Root + "/conf/dsl/" + project)
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to read /conf/dsl/"+project+", err="+err.Error())
			srg.EchoJSON(w, r, resDefault)
			return
		}
		//存储project的所有tag，为后面上报去重准备
		if _, ok := module.DSLModuleData[project]; !ok {
			module.DSLModuleData[project] = make(map[string]int)
		}
		for _, data := range contents {
			v := dslData{}
			err := json.Unmarshal([]byte(data), &v)
			if err != nil {
				tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to unmarshal "+data+", err="+err.Error())
				continue
			}
			if _, ok := module.DSLModuleData[project][v.Tag]; !ok {
				module.DSLModuleData[project][v.Tag] = 1
			}
			res = append(res, v)
		}
	}

	srg.EchoJSON(w, r, res)
}

//PlatformPostDsl nuwa平台接口 上报模块dsl数据
func (srg ShaRinGan) PlatformPostDsl(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	resErr := map[string]interface{}{"errno": 1, "errmsg": ""}
	resSuc := map[string]interface{}{"errno": 0, "errmsg": "success"}
	project := r.FormValue("project")
	tag := r.FormValue("tag")
	dsl := r.FormValue("dsl")
	user := r.FormValue("user")
	if user == "" {
		user = global.Sharingan
	}
	if project == "" || tag == "" || dsl == "" {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Params project or tag or dsl is illegal!")
		resErr["errmsg"] = "Params project or tag or dsl is illegal!"
		srg.EchoJSON(w, r, resErr)
		return
	}
	//优先读取http接口
	if url := conf.Handler.GetString("http_api.dsl_push"); url != "" {
		header := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
		_, bodyBytes, err := httpclient.Handler.Post(r.Context(), url, []byte("project="+project+"&tag="+tag+"&dsl="+dsl+"&user="+user), 0, header)
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to curl "+url+", "+err.Error())
			resErr["errmsg"] = err.Error()
			srg.EchoJSON(w, r, resErr)
			return
		}
		err = json.Unmarshal(bodyBytes, &resSuc)
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to unmarshal "+url+", "+err.Error())
			resErr["errmsg"] = err.Error()
			srg.EchoJSON(w, r, resErr)
			return
		}

	} else {
		//写入配置文件conf/dsl/{project}
		//写入配置前需要确认是否有重复tag
		if _, ok := module.DSLModuleData[project]; ok {
			if _, ok := module.DSLModuleData[project][tag]; ok {
				tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Error 1062: Duplicate tag "+tag+"!")
				resErr["errmsg"] = "Error 1062: Duplicate tag " + tag + "!"
				srg.EchoJSON(w, r, resErr)
				return
			}
		}
		//写入文件
		dslFD, err := os.OpenFile(conf.Root+"/conf/dsl/"+project, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		defer dslFD.Close()
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to open /conf/dsl/"+project+", err="+err.Error())
			resErr["errmsg"] = err.Error()
			srg.EchoJSON(w, r, resErr)
			return
		}
		data := map[string]interface{}{"dsl": dsl, "tag": tag, "project": project, "recommend": 1, "addTime": time.Now().Format("2006-01-02 15:04:05")}
		dataByte, _ := json.Marshal(data)
		_, err = dslFD.WriteString(string(dataByte) + "\n")
		if err != nil {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to write /conf/dsl/"+project+", err="+err.Error())
			resErr["errmsg"] = err.Error()
			srg.EchoJSON(w, r, resErr)
			return
		}
		//更新project的tag，为后面上报去重准备
		if _, ok := module.DSLModuleData[project]; !ok {
			module.DSLModuleData[project] = make(map[string]int)
		}
		module.DSLModuleData[project][tag] = 1
	}

	srg.EchoJSON(w, r, resSuc)
}

// CodeCoverage 代码覆盖率
func (srg ShaRinGan) CodeCoverage(w http.ResponseWriter, r *http.Request) {
	conf.FreshHandler()
	resErr := map[string]interface{}{"errno": 1, "errmsg": ""}
	resSuc := map[string]interface{}{"errno": 0, "errmsg": "success"}

	// !!!project值是 被测模块的bin文件名，或 bin文件名加 xx- 前缀, 如 platform  或 nuwa-platform!!!
	project := r.FormValue("project")
	if project == "" {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Params project is illegal!")
		resErr["errmsg"] = "Params project is illegal! Project should be same with moduleinfo.json or nuwa platform!"
		srg.EchoJSON(w, r, resErr)
		return
	}

	//读取project的服务地址，判断是否为本机(只支持 ShaRinGan与被测服务同机部署时，统计覆盖率)
	// listenAddr
	listenAddr := nuwaplt.GetValueByKey(project, nuwaplt.KListenAddr, "")
	if listenAddr == "" || !strings.Contains(listenAddr, "127.0.0.1") {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "ListenAddr of "+project+" is not 127.0.0.1:x!")
		resErr["errmsg"] = "ListenAddr of " + project + " is not 127.0.0.1:x! ShaRinGan and " + project + " should be deployed on the same machine!"
		srg.EchoJSON(w, r, resErr)
		return
	}

	// 判断project对应的 project.test服务是否启动(统计覆盖率的前提，以bin.test启动服务), 若存在，则读取 pid
	binName := project
	if strings.Contains(binName, "-") {
		tN := strings.Split(binName, "-")
		binName = tN[len(tN)-1]
	}
	stat, data := helper.GetPidByName(binName + ".test")
	if stat != 0 {
		resErr["errmsg"] = data
		if stat == 2 {
			tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, data)
			resErr["errmsg"] = data + " Please run with process '" + binName + ".test! "
		}
		srg.EchoJSON(w, r, resErr)
		return
	}
	pid := strings.TrimSpace(data)

	//读取完整的服务命令 (sudo vmmap 19781|grep 'Path:'  ------  ll /proc/6643/exe)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("ls -l /proc/%s/exe  | awk '{print $11}'", pid))
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf("ps -ef | grep '%s' | grep -v grep | grep -v '/bin/bash -c' | awk '{print $8}'", binName+".test"))
	}
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Errors happened when getting full path of "+binName+".test!")
		resErr["errmsg"] = "Errors happened when getting full path of " + binName + ".test!"
		srg.EchoJSON(w, r, resErr)
		return
	}
	fullEXE := strings.TrimSpace(helper.BytesToString(out))

	//kill被测服务，以生成.cov文件, 然后重启服务，保证后面测试正常（.cov文件以 起始yyyymmddHHiiss为后缀）
	//加锁
	srg.mu.Lock()
	defer srg.mu.Unlock()
	stat = helper.MkdirForce(global.DirCodeCov)
	if stat != 0 {
		resErr["errmsg"] = "Failed to mkdir " + global.DirCodeCov + ", please make it manually!"
		srg.EchoJSON(w, r, resErr)
		return
	}
	cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf("kill %s", pid))
	out, err = cmd.CombinedOutput()
	if err != nil || len(out) != 0 {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Errors happened when killing "+binName+".test! pid="+pid)
		resErr["errmsg"] = "Errors happened when killing " + binName + ".test! pid=" + pid
		srg.EchoJSON(w, r, resErr)
		return
	}
	// 确认进程是否真的被kill
	time.Sleep(500000000)
	stat, data = helper.GetPidByName(binName + ".test")
	if stat != 2 {
		time.Sleep(500000000)
		stat, data = helper.GetPidByName(binName + ".test")
		if stat != 2 {
			resErr["errmsg"] = "Failed to kill process " + binName + ".test! pid=" + pid + ". Please check it manually!"
			srg.EchoJSON(w, r, resErr)
			return
		}
	}
	//重启服务 （.cov文件以 起始yyyymmddHHiiss为后缀）
	fullCMD := "cd " + filepath.Dir(fullEXE) + " && nohup " + fullEXE + " -test.coverprofile=" + global.DirCodeCov + "/coverage." + binName + "." + time.Now().Format("20060102150405") + ".cov &"
	cmd = exec.Command("/bin/bash", "-c", fullCMD)
	err = cmd.Start()
	if err != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to restart "+binName+"!")
		resErr["errmsg"] = "Failed to restart " + binName + ".test! Please restart it manually! cmd=" + fullCMD
		srg.EchoJSON(w, r, resErr)
		return
	}
	go func(cmd *exec.Cmd, binName string) {
		defer func() {
			if err := recover(); err != nil {
				tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Panic happened when try to restart "+binName+"!")
			}
		}()

		err := cmd.Wait()
		if err != nil {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to restart "+binName+"! err="+err.Error())
		}
	}(cmd, binName)

	// 读取.cov文件名，并将.cov文件重命名，追加 结束yyyymmddHHiiss为后缀
	cmd = exec.Command("/bin/bash", "-c", "ls -trl "+global.DirCodeCov+"/coverage."+binName+"*.cov|tail -n 1|awk '{print $9}'")
	out, err = cmd.Output()
	if err != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to find "+global.DirCodeCov+"/coverage."+binName+"*.cov!")
		resErr["errmsg"] = "Failed to find " + global.DirCodeCov + "/coverage." + binName + "*.cov, please check it or try to replay again!"
		srg.EchoJSON(w, r, resErr)
		return
	}
	ori := strings.TrimSpace(helper.BytesToString(out))
	des := strings.TrimSuffix(ori, "cov") + time.Now().Format("20060102150405")
	stat = helper.MVFile(ori, des)
	if stat != 0 {
		des = ori
	}

	//使用gocov等开源工具，生成易读的覆盖率报告(报告里给出统计的起始时间范围)
	//gocov convert report/ut_coverage.out | gocov-html > report/ut_coverage.html
	// 读取.cov文件名，并将.cov文件重命名，追加 结束yyyymmddHHiiss为后缀
	covPath := conf.Root + "/install/codeCov/linux/"
	if runtime.GOOS == "darwin" {
		covPath = conf.Root + "/install/codeCov/darwin/"
	}
	covCmd := covPath + "gocov convert " + des + " | " + covPath + `gocov-html`
	// go 命令行不支持 重定向 >
	stdout, err := os.OpenFile(des+".html", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to create file "+des+".html! err="+err.Error())
		resErr["errmsg"] = "Failed to create file " + des + ".html! Please see original file " + des
		srg.EchoJSON(w, r, resErr)
		return
	}
	defer stdout.Close()
	cmd = exec.Command("/bin/bash", "-c", covCmd)
	cmd.Stdout = stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		tlog.Handler.Errorf(r.Context(), tlog.DLTagUndefined, "Failed to execute"+covCmd+"! err="+err.Error() + stderr.String())
		resErr["errmsg"] = "Failed to execute " + covCmd + "!"
		srg.EchoJSON(w, r, resErr)
		return
	}
	http.Redirect(w, r, "/coverage/report/"+strings.TrimPrefix(des, global.DirCodeCov+"/")+".html", 302)
	srg.EchoJSON(w, r, resSuc)
}

func (srg ShaRinGan) CodeCoverageReport(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conf.FreshHandler()
	resErr := map[string]interface{}{"errno": 1, "errmsg": ""}

	// 判断文件是否存在
	covFile := global.DirCodeCov + "/" + ps.ByName("covFile")
	if !helper.PathExists(covFile) {
		resErr["errmsg"] = "Failed to find " + covFile + "!"
		srg.EchoJSON(w, r, resErr)
		return
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(covFile)
	if err != nil {
		resErr["errmsg"] = "Failed to read " + covFile + "!"
		srg.EchoJSON(w, r, resErr)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	srg.Echo(w, r, data)
}

func getContentType(ctype string) string {
	typeSplits := strings.Split(ctype, ";")
	if len(typeSplits) > 1 {
		return typeSplits[0]
	}
	return ctype
}
