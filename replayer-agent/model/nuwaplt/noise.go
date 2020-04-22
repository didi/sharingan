package nuwaplt

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/global"
	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/common/handlers/httpclient"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/utils/helper"
)

type ReportResult struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}

type NoiseInfo struct {
	Id      int    `json:"id"`
	Uri     string `json:"uri"`
	Noise   string `json:"noise"`
	Project string `json:"project"`
	User    string `json:"user"`
	AddTime string `json:"addTime"`
}

func ReportNoise(project, uri, noise, user string) (result *ReportResult) {
	result = &ReportResult{1, "?"}

	//优先读取http接口上报噪音
	nuwaURL := conf.Handler.GetString("http_api.noise_push")
	if nuwaURL != "" {
		if user == "" {
			user = global.Sharingan
		}
		v := url.Values{}
		v.Set("project", project)
		v.Set("uri", uri)
		v.Set("noise", noise)
		v.Set("user", user)

		headers := map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}
		_, body, err := httpclient.Handler.Post(
			context.TODO(),
			nuwaURL,
			[]byte(v.Encode()),
			0,
			headers,
		)

		if err != nil {
			result.Errmsg = err.Error()
			return
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			result.Errmsg = err.Error()
			return result
		}
		return result
	} else {
		//写入配置文件(噪音不需要去重，上报就写入配置文件)
		noiseFD, err := os.OpenFile(conf.Root+"/conf/noise/"+project, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		defer noiseFD.Close()
		if err != nil {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to open /conf/noise/"+project+", err="+err.Error())
			result.Errmsg = err.Error()
			return
		}
		data := map[string]interface{}{"uri": uri, "noise": noise, "project": project, "addTime": time.Now().Format("2006-01-02 15:04:05")}
		dataByte, _ := json.Marshal(data)
		_, err = noiseFD.WriteString(string(dataByte) + "\n")
		if err != nil {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to write /conf/noise/"+project+", err="+err.Error())
			result.Errmsg = err.Error()
			return
		}

		return &ReportResult{0, "success"}
	}
}

////配置文件的噪音格式
//type noiseData struct {
//	Uri     string `json:"uri"`
//	Noise   string `json:"noise"`
//	Project string `json:"project"`
//	AddTime string `json:"addTime"`
//}

func DelNoise(id, project, uri, noise string) (result *ReportResult) {
	result = &ReportResult{1, "?"}

	//优先读取http接口删除噪音
	nuwaURL := conf.Handler.GetString("http_api.noise_del")
	if nuwaURL != "" {
		v := url.Values{}
		v.Set("id", id)
		headers := map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}
		_, body, err := httpclient.Handler.Post(
			context.TODO(),
			nuwaURL,
			[]byte(v.Encode()),
			0,
			headers,
		)

		if err != nil {
			result.Errmsg = err.Error()
			return
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			result.Errmsg = err.Error()
			return result
		}
		return result
	} else {
		//读取配置文件，找到要删除的噪音，删除后重新写回配置文件
		//读取配置文件conf/noise/{project}
		contents, err := helper.ReadLines(conf.Root + "/conf/noise/" + project)
		if err != nil {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to read /conf/noise/"+project+", err="+err.Error())
			result.Errmsg = err.Error()
			return result
		}
		//最终结果
		res := ""
		uniq := make(map[string]int)
		for _, data := range contents {
			v := NoiseInfo{}
			err := json.Unmarshal([]byte(data), &v)
			if err != nil {
				tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to unmarshal "+data+", err="+err.Error())
				continue
			}
			//命中删除的噪音
			if v.Uri != uri || v.Noise != noise {
				if _, ok := uniq[v.Uri+v.Noise]; !ok {
					uniq[v.Uri+v.Noise] = 1
					res = res + data + "\n"
				}
			}
		}
		//写回文件
		err = helper.WriteFileString(conf.Root+"/conf/noise/"+project, res)
		if err != nil {
			result.Errmsg = err.Error()
			return result
		}

		return &ReportResult{0, "success"}
	}
}
func GetNoise(project, uri string) (noises map[string]NoiseInfo) {
	noises = make(map[string]NoiseInfo)

	//优先读取http接口请求噪音数据
	nuwaURL := conf.Handler.GetString("http_api.noise_get")
	if nuwaURL != "" {
		v := url.Values{}
		v.Set("project", project)
		v.Set("uri", uri)
		_, body, httpErr := httpclient.Handler.Get(
			context.TODO(),
			nuwaURL+"?"+v.Encode(),
		)

		if httpErr != nil {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "errmsg=http get noise fail || err=%s", httpErr)
			return
		}

		var infos []NoiseInfo
		err := json.Unmarshal(body, &infos)

		if err == nil {
			for _, info := range infos {
				noises[info.Noise] = info
			}
		} else {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "errmsg=%s", err)
		}
	} else {
		//读取配置文件，过滤噪音
		//读取配置文件conf/noise/{project}
		contents, err := helper.ReadLines(conf.Root + "/conf/noise/" + project)
		if err != nil {
			tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to read /conf/noise/"+project+", err="+err.Error())
			return
		}
		for _, data := range contents {
			v := NoiseInfo{}
			err := json.Unmarshal([]byte(data), &v)
			if err != nil {
				tlog.Handler.Errorf(context.TODO(), tlog.DLTagUndefined, "Failed to unmarshal "+data+", err="+err.Error())
				continue
			}
			//命中要找的噪音
			if v.Uri == uri {
				noises[v.Noise] = v
			}
		}
	}
	return
}
