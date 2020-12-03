package request

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

// 请求实例
type request struct {
	uri url.URL
}

// New 初始化请求
func New(host string) *request {
	u := strings.Split(host, "://")
	scheme := "https"
	if len(u) > 1 {
		scheme = u[0]
	}
	return &request{uri: url.URL{
		Scheme: scheme,
		Host:   u[len(u)-1],
	}}
}

var client = http.DefaultClient

// protofile 请求文件结构
type file struct {
	data []byte // 文件数据
	name string
}

// Request 请求结构体
type Request struct {
	body   []byte
	params map[string]string
	header map[string]string
	file   *file
}

type option struct {
	request  *Request
	response interface{}
}

// Option 配置
type Option func(*option)

// WithBody 配置请求Body
func WithBody(body []byte) Option {
	return func(o *option) {
		o.request.body = body
	}
}

// WithHeader 配置请求Header
func WithHeader(k, v string) Option {
	return func(o *option) {
		o.request.header[k] = v
	}
}

// WithParam 配置请求Param
func WithParam(k, v string) Option {
	return func(o *option) {
		o.request.params[k] = v
	}
}

// WithFile 配置请求File
func WithFile(filename string, data []byte) Option {
	return func(o *option) {
		o.request.file = &file{name: filename, data: data}
	}
}

// WithResponse 配置响应结构体
func WithResponse(resp interface{}) Option {
	return func(o *option) {
		o.response = resp
	}
}

func (r *request) Do(method string, path string, opts ...Option) error {
	o := &option{
		request: &Request{
			params: make(map[string]string),
			header: make(map[string]string),
		},
	}
	for _, opt := range opts {
		opt(o)
	}
	params := url.Values{}
	for k, v := range o.request.params {
		params.Set(k, v)
	}
	r.uri.Path = path
	r.uri.RawQuery = params.Encode()
	var req *http.Request
	var err error
	if o.request.file != nil {
		// 上传文件请求
		buffer := new(bytes.Buffer)
		w := multipart.NewWriter(buffer)
		fw, err := w.CreateFormFile("media", o.request.file.name)
		if err != nil {
			return err
		}
		_, err = fw.Write(o.request.file.data)
		if err := w.Close(); err != nil {
			return err
		}
		req, err = http.NewRequest(method, r.uri.String(), buffer)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
	} else {
		req, err = http.NewRequest(method, r.uri.String(), bytes.NewReader(o.request.body))
		if err != nil {
			return err
		}
	}
	for k, v := range o.request.header {
		req.Header.Set(k, v)
	}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(rsp.Body).Decode(o.response); err != nil {
		return err
	}
	return nil
}

func (r *request) Get(path string, opts ...Option) error {
	return r.Do("GET", path, opts...)
}
func (r *request) Post(path string, opts ...Option) error {
	return r.Do("POST", path, opts...)
}
