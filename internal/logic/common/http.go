package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"lexa-engine/internal/logic"
)

type FormPayload struct {
	Payload  *bytes.Buffer `json:"payload"`
	FormType string        `json:"formType"`
}

type RequestQuery struct {
	QueryName  string `json:"queryName"`
	QueryValue string `json:"queryValue"`
	Type       string `json:"type"`
}
type RequestFormBodyParameter struct {
	FormName  string `json:"formName"`
	FormValue string `json:"formValue"`
}

type RequestHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Request struct {
	Method      string                     `json:"method"`
	ReqUrl      string                     `json:"reqUrl"`
	Parameters  []RequestQuery             `json:"parameters"`
	BodyParam   []RequestFormBodyParameter `json:"bodyParam"`
	Headers     []RequestHeader            `json:"headers"`
	PostType    string                     `json:"postType"`
	SetCookies  bool                       `json:"setCookies"`
	NeedCookies bool                       `json:"needCookies"`
}

type HttpClient struct {
	http.Client
	Cookies []*http.Cookie `json:"cookies"`
}

func (hc *HttpClient) SendRequest(req *Request) (resp *http.Response, err error) {
	switch req.Method {
	case "GET":
		return hc.get(req)
	case "POST":
		if req.PostType == "json" {
			return
		}
		if req.PostType == "form-data" {
			return hc.postForm(req)
		}
		if req.PostType == "x-www-form-urlencoded" {
			return hc.postUrlEncoded(req)
		}
		return
	default:
		{
			err = errors.New("Unsupported Method")
			return
		}
	}
}

func (hc *HttpClient) get(reqSpec *Request) (resp *http.Response, err error) {
	reqUrl := reqSpec.ReqUrl + "?"
	if len(reqSpec.Parameters) > 0 {
		for _, param := range reqSpec.Parameters {
			reqUrl += fmt.Sprintf("%s=%s", param.QueryName, param.QueryValue)
		}
	}
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		logx.Error("构建/GET请求失败")
		return
	}
	if reqSpec.NeedCookies && len(hc.Cookies) > 0 {
		for _, c := range hc.Cookies {
			req.AddCookie(c)
		}
		logx.Debug("设置cookies成功")
	}

	if len(reqSpec.Headers) > 0 {
		for _, h := range reqSpec.Headers {
			req.Header.Add(h.Key, h.Value)
		}
	}
	resp, err = hc.Client.Do(req)
	if err != nil {
		logx.Error(fmt.Sprintf("发送http请求失败 url=[%v]", reqUrl))
		return
	}
	if resp.StatusCode != logic.HTTP_OK_STATUS {
		err = logic.HTTP_STATUS_NOT_200
		logx.Error(fmt.Sprintf("调用 [%s] 失败, 状态码=[%v]\n", reqSpec.ReqUrl, resp.StatusCode))
		logx.Error("返回体: ", string(ReadBodyByte(resp)))
		return
	}

	if reqSpec.SetCookies {
		hc.Cookies = resp.Cookies()
	}
	return
}

func (hc *HttpClient) postUrlEncoded(reqSpec *Request) (resp *http.Response, err error) {
	payloadString := ""
	for _, field := range reqSpec.BodyParam {
		payloadString += fmt.Sprintf("%s=%s&", field.FormName, field.FormValue)
	}
	payloadString = strings.TrimRight(payloadString, "&")
	logx.Error(payloadString)
	payload := strings.NewReader(payloadString)
	req, err := http.NewRequest("POST", reqSpec.ReqUrl, payload)
	if err != nil {
		logx.Error(err)
		return
	}

	req.Header.Add("Client-Type", "PC")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = hc.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 状态码不为200
	if resp.StatusCode != logic.HTTP_OK_STATUS {
		err = logic.HTTP_STATUS_NOT_200
		logx.Error(fmt.Sprintf("调用 [%s] 失败, 状态码=[%v]\n", reqSpec.ReqUrl, resp.StatusCode))
		logx.Error("返回体: ", string(ReadBodyByte(resp)))
		return
	}

	// 需要保存cookie
	if reqSpec.SetCookies {
		hc.Cookies = resp.Cookies()
	}

	return
}

func (hc *HttpClient) postForm(reqSpec *Request) (resp *http.Response, err error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	for _, field := range reqSpec.BodyParam {
		err = writer.WriteField(string(field.FormName), string(field.FormValue))
		if err != nil {
			logx.Error(err)
			return
		}
	}

	if err = writer.Close(); err != nil {
		logx.Error(err)
		return
	}

	request, err := http.NewRequest("POST", reqSpec.ReqUrl, payload)
	if err != nil {
		logx.Error(err)
		return
	}

	if reqSpec.NeedCookies && len(hc.Cookies) > 0 {
		for _, c := range hc.Cookies {
			request.AddCookie(c)
		}
		logx.Debug("设置cookies成功")
	}

	if len(reqSpec.Headers) > 0 {
		for _, h := range reqSpec.Headers {
			request.Header.Add(h.Key, h.Value)
		}
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err = hc.Client.Do(request)
	if err != nil {
		logx.Error(err)
		return
	}

	// 状态码不为200
	if resp.StatusCode != logic.HTTP_OK_STATUS {
		err = logic.HTTP_STATUS_NOT_200
		logx.Error(fmt.Sprintf("调用 [%s] 失败, 状态码=[%v]\n", reqSpec.ReqUrl, resp.StatusCode))
		logx.Error("返回体: ", string(ReadBodyByte(resp)))
		return
	}

	// 需要保存cookie
	if reqSpec.SetCookies {
		hc.Cookies = resp.Cookies()
	}
	return
}

func ReadBody(resp *http.Response) (bodyMap map[string]any, err error) {
	defer resp.Body.Close()

	// 读取response,转换成map
	bodyByte, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		err = readErr
		logx.Error(err)
		return
	}
	bodyMap = make(map[string]any)
	if err = json.Unmarshal(bodyByte, &bodyMap); err != nil {
		logx.Error("读取response到map失败\n", err)
		return
	}
	return
}

func ReadBodyByte(resp *http.Response) (body []byte) {
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logx.Error("读取response.body失败 \n", readErr)
		return
	}
	return
}
