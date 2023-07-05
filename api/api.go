package api

import (
	"encoding/json"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"mj-wechat-bot/errorhandler"
	"net/http"
	"net/url"
	"time"
)

type API struct {
	ApiUrl   string `yaml:"api_url"`
	Apikey   string `yaml:"api_key"`
	CheckUrl string `yaml:"check_url"`
	TianqiApi string `yaml:"tianqi_api"`
	TianqiAppCode string `yaml:"tianqi_appCode"`
	StoryApi string `yaml:"story_api"`
}

var config API
var (
	createUrl     string
	taskUrl       string
	taskUpdateUrl string
	TianqiApi 	  string
	StoryApi 	  string
)

func init() {
	// 注册异常处理函数
	defer errorhandler.HandlePanic()
	// Read configuration file.
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %v", err))
	}

	// Unmarshal configuration.

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}
	createUrl = config.ApiUrl + "/submit/imagine"
	taskUrl = config.ApiUrl + "/task"
	taskUpdateUrl = config.ApiUrl + "/submit/simple-change"
	TianqiApi = config.TianqiApi
	StoryApi = config.StoryApi

}

type Response struct {
	Code int                    `json:"code"`
	Task string		    		`json:"result"`
	Data map[string]interface{} `json:"properties"`
	Msg  string                 `json:"description"`
	Time int                    `json:"time"`
}

type TaskResponse struct {
	Action string                  `json:"action"`
	Task string		    `json:"id"`
	Data map[string]interface{} `json:"properties"`
	Msg  string                 `json:"description"`
	Status  string              `json:"status"`
	ImageUrl  string            `json:"imageUrl"`
	Progress  string            `json:"progress"`
	Prompt    string            `json:"prompt"`
	PromptEn  string            `json:"promptEn"`
	StartTime  int		    `json:"startTime"`
	FinishTime int		    `json:"finishTime"`
	SubmitTime int		    `json:"submitTime"`
	FailReason string           `json:"failReason"`
}

type TianqiResponse struct {
	Status int                   `json:"status"`
	Msg string		    			`json:"msg"`
	Result map[string]interface{} 	`json:"result"`
}

type StoryResponse struct {
	Code int                   		`json:"code"`
	Msg  string		    			`json:"msg"`
	Data map[string]interface{} 	`json:"data"`
}

func CreateMessage(text string) (bool, string) {
	reqUrl, err := url.Parse(createUrl)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	query := `{"prompt":"` + text + `"}`

	body, err := DoPost(reqUrl,query)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	var response Response
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		fmt.Println(err)
		return false, ""
	}
	if response.Code != 1 {
		fmt.Println(response.Msg)
		return false, ""
	}
	return true, response.Task
}

//查询任务状态
func QueryTaskStatus(taskID string) (bool, map[string]interface{}) {
	reqUrl, err := url.Parse(taskUrl + "/" + taskID + "/fetch")
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	query := reqUrl.Query()
	//query.Set("task_id", taskID)
	reqUrl.RawQuery = query.Encode()

	body, err := DoGet(reqUrl,nil)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	log.Printf("任务【%s】返回结果 -> %s", taskID, body)
	var response TaskResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		fmt.Println(err)
		return false, nil
	}
	//if response.Status == "FAILURE" {
	//	fmt.Println(response.Msg)
	//	return false, nil
	//}

	res := make(map[string]interface{})
	res["status"] = response.Status
	res["image_url"] = response.ImageUrl
	res["fail_reason"] = response.FailReason

	return true, res
}

//查询天气
func QueryTianqi(city string) (bool, map[string]interface{}) {
	reqUrl, err := url.Parse(TianqiApi)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	query := reqUrl.Query()
	query.Set("city", city)
	reqUrl.RawQuery = query.Encode()


	headers := make(map[string]string)
	headers["Authorization",] = "APPCODE "+config.TianqiAppCode

	body, err := DoGet(reqUrl,headers)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	log.Printf("城市天气【%s】返回结果 -> %s", city, body)
	var response TianqiResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		fmt.Println(err)
		return false, nil
	}
	if response.Status != 0 {
		fmt.Println(response.Msg)
		return false, nil
	}

	return true, response.Result
}

//查询故事，可以传入故事标题
func QueryStory(title string) (bool, map[string]interface{}) {
	reqUrl, err := url.Parse(StoryApi)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	query := reqUrl.Query()
	query.Set("title", title)
	reqUrl.RawQuery = query.Encode()


	headers := make(map[string]string)
	headers["Authorization",] = "APPCODE "+config.StoryApi

	body, err := DoGet(reqUrl,headers)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	if err != nil {
		fmt.Println(err)
		return false, nil
	}

	log.Printf("听故事【%s】返回结果 -> %s", title, body)
	var response StoryResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		fmt.Println(err)
		return false, nil
	}
	if response.Code != 1 {
		fmt.Println(response.Msg)
		return false, nil
	}

	return true, response.Data
}

func TaskUpdate(taskId string, action string) (bool, string) {
	reqUrl, err := url.Parse(taskUpdateUrl)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}

	log.Printf("reqUrl: %s", reqUrl.String())

	query := `{"content":"` + taskId + ` `+action+`"}`
	body, err := DoPost(reqUrl,query)

	if err != nil {
		fmt.Println(err)
		return false, ""
	}

	var response Response
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		fmt.Println(err)
		return false, ""
	}
	if response.Code != 1 {
		fmt.Println(response.Msg)
		return false, ""
	}
	return true, response.Task
}

func DoGet(reqUrl *url.URL,headers map[string]string) (string, error) {
	// 构建 HTTP GET 请求
	req, err := http.NewRequest("GET", reqUrl.String(), nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// 创建一个 HTTP 客户端
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	// 添加请求头
	if headers != nil {
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}else{
		req.Header.Add("mj-api-secret", config.Apikey)
	}

	// 发送 HTTP GET 请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}

func DoPost(reqUrl *url.URL, msg string) (string, error) {

	postData := bytes.NewBuffer([]byte(msg))

	// 构建 HTTP POST 请求
	req, err := http.NewRequest("POST", reqUrl.String(), postData)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// 创建一个 HTTP 客户端
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	// 添加请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("mj-api-secret", config.Apikey)
	// 发送 HTTP Post 请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}