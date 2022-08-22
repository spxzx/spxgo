package spxgo

import (
	"fmt"
	"gitbuh.com/spxzx/spxgo/config"
	spxLog "gitbuh.com/spxzx/spxgo/log"
	"gitbuh.com/spxzx/spxgo/render"
	"html/template"
	"log"
	"net/http"
	"sync"
)

const MethodAny = "ANY"

type HandlerFunc func(c *Context)

// MiddlewareFunc 传入 HandlerFunc 处理完后再将其返回 ->达成影响代码
type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

type routerGroup struct {
	name               string
	treeNode           *treeNode
	handlerFuncMap     map[string]map[string]HandlerFunc      // { "name": { "method": HandleFunc } }
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc // { "name": { "method": []MiddlewareFunc } }
	middlewares        []MiddlewareFunc                       // 通用中间件
}

type router struct {
	routerGroups []*routerGroup
	engine       *Engine
}

type ErrorHandler func(err error) (int, any)

// Use 可能加入多个
func (r *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewareFunc...)
}

// 中间件的处理
func (r *routerGroup) methodHandle(name string, method string, handleFunc HandlerFunc, c *Context) {
	// 组通用中间件 preMiddlewares ** 废弃 已改成下面代码
	if r.middlewares != nil {
		for _, middlewareFunc := range r.middlewares {
			handleFunc = middlewareFunc(handleFunc) // 覆盖处理
		}
	}
	// 组路由级中间件
	if middlewareFuncs := r.middlewaresFuncMap[name][method]; middlewareFuncs != nil {
		for _, middlewareFunc := range middlewareFuncs {
			handleFunc = middlewareFunc(handleFunc) // 覆盖处理
		}
	}
	handleFunc(c) // 真正执行   pre -> handle() <- post
	// postMiddlewares ** 废弃
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:               name,
		treeNode:           &treeNode{name: "/", children: make([]*treeNode, 0)},
		handlerFuncMap:     make(map[string]map[string]HandlerFunc),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		middlewares:        make([]MiddlewareFunc, 0),
	}
	routerGroup.Use(r.engine.middles...)
	r.routerGroups = append(r.routerGroups, routerGroup)
	return routerGroup
}

// router1 get->handle
// router2 post->handle
func (r *routerGroup) handle(name string, method string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	if _, ok := r.handlerFuncMap[name]; !ok {
		r.handlerFuncMap[name] = make(map[string]HandlerFunc)
		r.middlewaresFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	if _, ok := r.handlerFuncMap[name][method]; ok {
		panic("there are duplicate routes")
	}
	r.handlerFuncMap[name][method] = handlerFunc
	r.middlewaresFuncMap[name][method] = append(r.middlewaresFuncMap[name][method], middlewareFunc...)
	r.treeNode.Put(name)
}

func (r *routerGroup) Any(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, MethodAny, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Get(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodGet, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Post(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPost, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Delete(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodDelete, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Put(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPut, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Patch(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodPatch, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Options(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodOptions, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Head(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, http.MethodHead, handlerFunc, middlewareFunc...)
}

type Engine struct {
	router
	funcMap      template.FuncMap // template.FuncMap 存疑
	HTMLRender   render.HTML
	pool         sync.Pool      // sync.Pool 用于存储那些被分配了但是还没有被使用，但是未来可能使用的值，这样可以不用再次分配内存，提高效率
	Logger       *spxLog.Logger // 分级日志
	middles      []MiddlewareFunc
	errorHandler ErrorHandler
}

func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}

func New() *Engine {
	engine := &Engine{
		// router: router{handleFuncMap: make(map[string]HandleFunc)},
		router: router{},
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) Use(middlewareFunc ...MiddlewareFunc) {
	e.middles = append(e.middles, middlewareFunc...)
}

func Default() *Engine {
	engine := New()
	engine.Logger = spxLog.Default()
	logPath, ok := config.Conf.Log["path"]
	if ok {
		engine.Logger.SetLogPath(logPath.(string))
	}
	engine.Use(Logging, Recovery)
	engine.router.engine = engine // 使能够使用中间件
	return engine
}

// SetFuncMap
// Deprecated
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap // 设置 map[string]any 映射键值对
}

func (e *Engine) SetHTMLRender(t *template.Template) {
	e.HTMLRender = render.HTML{Template: t}
}

func (e *Engine) LoadTemplate(pattern string) {
	// New() 创建一个名为name的模板 Funcs() 模板template的函数字典里加入参数funcMap内的键值对
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHTMLRender(t)
}

// 具体处理请求，匹配ANY,GET,POST等请求方法，匹配中间件
func (e *Engine) httpRequestHandle(c *Context, w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, group := range e.routerGroups {
		// URL不能使用r.RequestURI,这个会包含传来的参数
		routerName := subStringLast(r.URL.Path, "/"+group.name)
		// routerName /get/1   mode /get/:id
		node := group.treeNode.Get(routerName)
		if node != nil && node.isLeaf {
			// 路由匹配成功
			if handlerFunc, ok := group.handlerFuncMap[node.routerName][MethodAny]; ok {
				group.methodHandle(node.routerName, MethodAny, handlerFunc, c)
				return
			}
			if handlerFunc, ok := group.handlerFuncMap[node.routerName][method]; ok {
				group.methodHandle(node.routerName, method, handlerFunc, c)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = fmt.Fprintf(w, "%s %s not allowed \n", r.RequestURI, method)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "%s not found \n", r.RequestURI)
}

// 实现http下的接口ServeHTTP
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// pool -> 为了解决频繁创建Context的问题
	c := e.pool.Get().(*Context)
	c.W = w
	c.R = r
	c.Logger = e.Logger
	e.httpRequestHandle(c, w, r)
	e.pool.Put(c)
}

func (e *Engine) RegisterErrorHandler(handler ErrorHandler) {
	e.errorHandler = handler
}

func (e *Engine) Run(addr string) {
	// routerX{ group_name   HF{ key: get value: func } }
	/*for _, group := range e.routerGroups {
		for key, value := range group.handleFuncMap {
			http.HandleFunc("/"+group.name+key, value)
		}
	}*/
	// 使所有请求进入到ServeHTTP进行判断 -> ServeHTTP(w http.ResponseWriter, r *http.Request)
	http.Handle("/", e)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) Handler() http.Handler {
	return e
}

func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	err := http.ListenAndServeTLS(addr, certFile, keyFile, e.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
