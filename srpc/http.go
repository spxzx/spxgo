package sprc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type HttpClient struct {
	client     http.Client
	serviceMap map[string]Service
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: http.Client{
			Timeout: time.Duration(3) * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   5,
				MaxConnsPerHost:       100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		serviceMap: make(map[string]Service),
	}
}

func (c *HttpClient) responseHandle(request *http.Request) ([]byte, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		info := fmt.Sprintf("response status is %d", response.StatusCode)
		return nil, errors.New(info)
	}
	buf := make([]byte, 79)
	body := make([]byte, 0)
	reader := bufio.NewReader(response.Body)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF || n == 0 {
			break
		}
		body = append(body, buf[:n]...)
		if n < len(buf) {
			break
		}
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Print(err)
		}
	}(response.Body)
	return body, nil
}

func (c *HttpClient) Response(r *http.Request) ([]byte, error) {
	return c.responseHandle(r)
}

func (c *HttpClient) toValues(args map[string]any) string {
	if args != nil && len(args) > 0 {
		params := url.Values{}
		for k, v := range args {
			params.Set(k, fmt.Sprintf("%v", v))
		}
		return params.Encode()
	}
	return ""
}

func (c *HttpClient) GetRequest(url string, args map[string]any) (*http.Request, error) {
	if args != nil && len(args) > 0 {
		url += "?" + c.toValues(args)
	}
	return http.NewRequest("GET", url, nil)
}

func (c *HttpClient) Get(url string, args map[string]any) ([]byte, error) {
	if args != nil && len(args) > 0 {
		url += "?" + c.toValues(args)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *HttpClient) FormRequest(method string, url string, args map[string]any) (*http.Request, error) {
	return http.NewRequest(method, url, strings.NewReader(c.toValues(args)))

}

func (c *HttpClient) PostForm(url string, args map[string]any) ([]byte, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *HttpClient) JsonRequest(method string, url string, args map[string]any) (*http.Request, error) {
	marshal, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(method, url, bytes.NewReader(marshal))
}

func (c *HttpClient) PostJson(url string, args map[string]any) ([]byte, error) {
	marshal, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

type HttpConfig struct {
	Protocol string
	Host     string
	Port     string
	Ssl      bool
}

const (
	HTTP  = "HTTP"
	HTTPS = "HTTPS"
)

const (
	GET      = "GET"
	PostForm = "POST_FORM"
	PostJson = "POST_JSON"
)

type Service interface {
	Env() HttpConfig
}

func (c HttpConfig) ProtocolSelect() string {
	if c.Protocol == "" {
		c.Protocol = HTTP
	}
	switch c.Protocol {
	case HTTP:
		return "http://" + c.Host + ":" + c.Port
	case HTTPS:
		return "https://" + c.Host + ":" + c.Port
	}
	return ""
}

func (c *HttpClient) RegisterHttpService(name string, service Service) {
	c.serviceMap[name] = service
}

func (c *HttpClient) Do(serviceName string, method string) Service {
	service, ok := c.serviceMap[serviceName]
	if !ok {
		panic(errors.New("service not found"))
	}
	t, v := reflect.TypeOf(service), reflect.ValueOf(service)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("service not pointer"))
	}
	tVal, vVal := t.Elem(), v.Elem()
	fieldIndex := -1
	for i := 0; i < tVal.NumField(); i++ {
		if tVal.Field(i).Name == method {
			fieldIndex = i
			break
		}
	}
	if fieldIndex == -1 {
		panic(errors.New("method not found"))
	}
	tag := tVal.Field(fieldIndex).Tag
	rpcInfo := tag.Get("srpc")
	if rpcInfo == "" {
		panic(errors.New("not srpc tag"))
	}
	split := strings.Split(rpcInfo, ",")
	if len(split) != 2 {
		panic(errors.New("tag srpc not valid"))
	}
	methodType, path := split[0], split[1]
	httpConfig := service.Env()
	function := func(args map[string]any) ([]byte, error) {
		switch methodType {
		case GET:
			return c.Get(httpConfig.ProtocolSelect()+path, args)
		case PostForm:
			return c.PostForm(httpConfig.ProtocolSelect()+path, args)
		case PostJson:
			return c.PostJson(httpConfig.ProtocolSelect()+path, args)
		default:
			return nil, errors.New("no match method type")
		}
	}
	vVal.Field(fieldIndex).Set(reflect.ValueOf(function))
	return service
}
