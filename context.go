package spxgo

import (
	"errors"
	"fmt"
	"gitbuh.com/spxzx/spxgo/binding"
	"gitbuh.com/spxzx/spxgo/internal/bytesconv"
	spxLog "gitbuh.com/spxzx/spxgo/log"
	"gitbuh.com/spxzx/spxgo/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const (
	defaultMaxMemory       = 32 << 20 // 32M
	defaultMultipartMemory = 32 << 20
)

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	queryCache            url.Values // go提供了原生query参数的map支持
	postFormCache         url.Values
	StatusCode            int
	DisallowUnknownFields bool // 开启结构体中没有该属性 没有就报错 ! 但是如果传来的参数中有结构体中也没有的 _不会报错_ !
	IsValidate            bool // 开启传参中含有结构体中没有的属性的报错
	Logger                *spxLog.Logger
	Keys                  map[string]any // 认证信息
	mutex                 sync.RWMutex
	sameSite              http.SameSite // 为了做安全性操作
}

func (c *Context) Set(key string, value any) {
	c.mutex.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
	c.mutex.Unlock()
}

func (c *Context) Get(key string) (value any, ok bool) {
	c.mutex.Lock()
	value, ok = c.Keys[key]
	c.mutex.Unlock()
	return
}

func (c *Context) SetBasicAuth(username, password string) {
	c.R.Header.Set("Authorization", "Basic "+BasicAuth(username, password))
}

func (c *Context) SetSameSite(sameSite http.SameSite) {
	c.sameSite = sameSite
}

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.W, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: c.sameSite,
	})
}

// ==================== render 渲染 ========================

func (c *Context) Render(statusCode int, r render.Render) error {
	// WriteHeader 不能多次调用 否则会产生 http: superfluous response.WriteHeader 错误
	// 如果设置了 statusCode 对header的修改就不生效了。。。
	// c.W.WriteHeader(statusCode)
	err := r.Render(c.W, statusCode)
	c.StatusCode = statusCode
	return err
}

// HTML 单纯通过传过来的内容进行渲染，而后面的都是用已有的html模板去进行页面渲染
func (c *Context) HTML(status int, data string) error {
	return c.Render(status, &render.HTML{Data: data, IsTemplate: false})
}

// region

// HTMLTemplateFiles ** 废弃功能
// Deprecated
func (c *Context) HTMLTemplateFiles(name string, data any, filenames ...string) error {
	// 设置状态 默认不设置的话 如果调用了write这个方法 实际上默认返回200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	// 执行
	err = t.Execute(c.W, data)
	return err
}

// HTMLTemplateGlob ** 废弃功能
// Deprecated
func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	// 执行
	err = t.Execute(c.W, data)
	return err
}

// endregion

func (c *Context) Template(status int, name string, data any) error {
	return c.Render(status, &render.HTML{
		Name:       name,
		Data:       data,
		Template:   c.engine.HTMLRender.Template,
		IsTemplate: true,
	})
}

func (c *Context) JSON(status int, data any) error {
	return c.Render(status, &render.JSON{Data: data})
}

func (c *Context) XML(status int, data any) error {
	return c.Render(status, &render.XML{Data: data})
}

// Redirect 简单重定向
func (c *Context) Redirect(statusCode int, location string) error {
	return c.Render(statusCode, &render.Redirect{
		Request:  c.R,
		Location: location,
	})
}

// StringOld 重构的示例
// Deprecated
func (c *Context) StringOld(status int, format string, values ...any) error {
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(status)
	if len(values) > 0 {
		_, err := fmt.Fprintf(c.W, format, values...)
		return err
	}
	_, err := c.W.Write(bytesconv.StringToBytes(format))
	return err
}

func (c *Context) String(status int, format string, values ...any) error {
	return c.Render(status, &render.String{Format: format, Data: values})
}

// File 文件
func (c *Context) File(filename string) {
	http.ServeFile(c.W, c.R, filename)
}

// FileAttachment 文件下载支持文件名
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition",
			`attachment; filename="`+filename+`"`)
	} else {
		// QueryEscape函数对s进行转码使之可以安全的用在URL查询里,解析后符号变为%AB形式,用于存储诸如中文为名等形式的文件
		c.W.Header().Set("Content-Disposition",
			`attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// FileFromFS filepath 是相对文件系统的路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		// 将原有的path进行恢复
		c.R.URL.Path = old
	}(c.R.URL.Path) // 这个值是在未赋下面值之前传入的 即请求路由的原始路径 而非文件路径
	c.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

// ====================参数解析=============================

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) QueryArray(key string) []string {
	c.initQueryCache()
	values, _ := c.GetQueryArray(key)
	return values
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) GetDefaultQuery(key, defaultValue string) string {
	values, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return values[0]
}

func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	for k, v := range m { // m : { k->queryKey: v->[]queryValues }
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				// k[i+1:][:j] <== (k[i+1:])[:j]
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

// GetQueryMap ....?usr[id]=1&usr[id]=2&usr[name]=spxzx
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) QueryMap(key string) map[string]string {
	dicts, _ := c.GetQueryMap(key)
	return dicts
}

func (c *Context) initPostFormCache() {
	if c.R != nil {
		if err := c.R.ParseMultipartForm(defaultMaxMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				c.Logger.Info(err)
			}
		}
		c.postFormCache = c.R.PostForm
	} else {
		c.postFormCache = url.Values{}
	}
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initPostFormCache()
	values, ok := c.postFormCache[key]
	return values, ok
}

func (c *Context) PostFormMap(key string) map[string]string {
	m, _ := c.GetPostFormMap(key)
	return m
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initPostFormCache()
	return c.get(c.postFormCache, key)
}

func (c *Context) MultipartFormFiles() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err // MultipartForm 是一个Form结构体 .File 是具体的文件map
}

// FormFile
// Deprecated
func (c *Context) FormFile(key string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(key) // 解析
	if err != nil {
		log.Println(err)
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)
	return header
}

func (c *Context) FormFiles(key string) []*multipart.FileHeader {
	multipartForm, err := c.MultipartFormFiles()
	if err != nil {
		log.Println(err)
		return []*multipart.FileHeader{}
	}
	return multipartForm.File[key]
}

func (c *Context) SaveUploadFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Println(err)
		}
	}(src)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Println(err)
		}
	}(out)
	_, err = io.Copy(out, src)
	return err
}

func (c *Context) SaveAllUploadFiles(files []*multipart.FileHeader, dst string) error {
	for _, file := range files {
		err := c.SaveUploadFile(file, dst+file.Filename)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func (c *Context) MustBindWith(obj any, bind binding.Binding) error {
	if err := c.ShouldBindWith(obj, bind); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) ShouldBindWith(obj any, bind binding.Binding) error {
	return bind.Bind(c.R, obj)
}

// BindJson 解析传参中的json数据
func (c *Context) BindJson(obj any) error {
	json := binding.JSON
	json.IsValidate = c.IsValidate
	json.DisallowUnknownFields = c.DisallowUnknownFields
	return c.MustBindWith(obj, json)
}

func (c *Context) BindXML(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}

// ======================

func (c *Context) Fail(statusCode int, msg string) {
	err := c.String(statusCode, msg)
	if err != nil {
		return
	}
}
