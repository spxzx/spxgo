# spxgo

一个为了提升自己水平而写的Go框架。

[TOC]



## binding

### binding.go

#### 常量

| 常量名 | 常量值        | 说明                   |
| ------ | ------------- | ---------------------- |
| JSON   | jsonBinding{} | 用于调用绑定JSON绑定器 |
| XML    | xmlBinding{}  | 用于调用绑定XML绑定器  |

#### 接口

Binding

| 方法                           | 说明                 |
| ------------------------------ | -------------------- |
| Name() string                  | 返回绑定器名字       |
| Bind(*http.Request, any) error | 绑定器绑定功能的实现 |

### json.go

#### 数据结构

jsonBinding

| 属性                  | 类型 | 说明                                                         |
| --------------------- | ---- | ------------------------------------------------------------ |
| IsValidate            | bool | 是否开启处理结构体中必要的属性（通过在属性后使用 \`spxgo:"required"\` 这个 tag 开启） |
| DisallowUnknownFields | bool | 是否开启处理参数中有而结构体中没有属性                       |

#### 方法/函数

`func (jsonBinding) Name() string`
实现Binding接口，返回绑定器的名字；

`func (b jsonBinding) Bind(r *http.Request, obj any) error`
实现Binding接口，将请求的json数据经过验证后解析并绑定回obj，使obj带有具体的json数据以方便解析渲染等操作；
decoder := json.NewDecoder(r.Body) 使解码器带有请求数据；
在最后调用了validate(obj)进行第三方验证；

`func checkParamDecoder(vOf reflect.Value, obj any, decoder *json.Decoder) error`
对单个数据结构类型的请求数据进行解析绑定；

`func checkParamSliceDecoder(elem reflect.Type, obj any, decoder *json.Decoder) error`
对切片、数组类型（内部有对应数据结构）的请求数据进行解析绑定；

`func validateParam(obj any, decoder *json.Decoder) error`
区分obj是 指针、结构体、切片、数组中的类型后调用checkParamDecoder()/checkParamSliceDecoder()进行具体的解析绑定

### validator.go

#### 全局变量

| 变量名    | 变量类型        | 变量值              | 说明                             |
| --------- | --------------- | ------------------- | -------------------------------- |
| Validator | StructValidator | &defaultValidator{} | 本框架提供一个默认的第三方验证器 |

#### 接口

**StructValidator**

第三方验证器的接口实现

| 方法                      | 说明                                       |
| ------------------------- | ------------------------------------------ |
| ValidateStruct(any) error | 验证的规范实现                             |
| Engine() any              | 有地方需要用到该验证器可以用这个返回其单例 |

#### 数据结构

defaultValidator

| 属性     | 类型                | 说明                                                      |
| -------- | ------------------- | --------------------------------------------------------- |
| one      | sync.Once           | 单例，一次是一个将执行一个动作的对象                      |
| validate | *validator.Validate | 第三方验证器validator，基于标签实现结构和单个字段的值验证 |

#### 自定义类型

| 类型名               | 具体类型 | 说明                    |
| -------------------- | -------- | ----------------------- |
| SliceValidationError | []error  | 用于一次性处理多个error |

#### 方法/函数

`func (err SliceValidationError) Error() string`
重写了error接口下的Errot方法，可以一次性处理多个错误；

`func (d *defaultValidator) lazyInit()`
懒加载第三方validator验证器；

`func (d *defaultValidator) validateStruct(obj any) error`
调用第三方validator验证器执行验证；

`func (d *defaultValidator) ValidateStruct(obj any) error`
实现接口，区分obj是 指针、结构体、切片、数组中的类型 后调用第三方validator验证器执行验证；

`func (d *defaultValidator) Engine() any`
实现接口，返回validator单例；

`func validate(obj any) error`
调用ValidateStruct()；本包下其它的go文件中可以通过该函数使用第三方验证器验证obj数据；

### xml.go

#### 数据结构

xmlBinding [暂] 无属性

#### 方法/接口

`func (xmlBinding) Name() string`
实现Binding接口，返回绑定器名字；

`func (b xmlBinding) Bind(r *http.Request, obj any) error`
实现Binding接口，将请求的xml数据经过验证后解析并绑定回obj，使obj带有具体的xml数据以方便解析渲染等操作；

`func decodeXML(r io.Reader, obj any) error`
进行具体的解析绑定，在最后调用了validate(obj)进行第三方验证；

------

## config

提供toml文件支持

### config.go

#### 全局变量

| 变量名 | 变量值                               | 说明                 |
| ------ | ------------------------------------ | -------------------- |
| Conf   | &SpxConfig{logger: spxLog.Default()} | 对外提供一个Conf配置 |

#### 数据结构

SpxConfig

| 属性   | 类型           | 说明                                                         |
| ------ | -------------- | ------------------------------------------------------------ |
| logger | *spxLog.Logger | 本框架的分级Log                                              |
| Log    | map[string]any | 配置文件[log] 日志配置<br />path="日志打印地址"              |
| Pool   | map[string]any | 配置文件[pool] 协程池配置<br />cap=协程池容量<br />expire=worker过期时间 |
| Mysql  | map[string]any | 配置文件[mysql] mysql数据库配置<br />username="用户名" <br />password="密码" <br />host="IP" <br />port="端口号" <br />dbName="数据库名字" |

#### 方法/函数

`func init()`
init()函数会在包被初始化后自动执行，并且在main()函数之前执行，init()无法被显示调用；
这里init会调用loadToml对.toml配置文件进行加载；

`func loadToml()`
加载.toml配置文件，[规范]默认位置以及文件名在： 项目名/conf/app.toml

------

## internal

### bytesconv

#### bytresconv.go

`func StringToBytes(s string) []byte`
将字符串 s 转换位字节数组；不会发生内存拷贝，性能相对较好；

### spxstings

#### spxstrings.go

`func JoinStrings(data ...any) string`
拼接字符串；

`func check(v any) string`
判断v的类型，若是string则将其断言v.(string)并返回，若不是则默认格式化字符串并返回；

------

## log

### log.go

#### 常量

| 常量名    | 常量值                           |
| --------- | -------------------------------- |
| greenBg   | 一系列颜色值，带Bg的改变背景颜色 |
| whiteBg   |                                  |
| yellowBg  |                                  |
| redBg     |                                  |
| blueBg    |                                  |
| magentaBg |                                  |
| cyanBg    |                                  |
| green     |                                  |
| white     |                                  |
| yellow    |                                  |
| red       |                                  |
| blue      |                                  |
| magenta   |                                  |
| cyan      |                                  |
| reset     | 颜色清空                         |

#### 自定义类型

| 类型名      | 具体类型       | 说明                     |
| ----------- | -------------- | ------------------------ |
| LoggerLevel | int            | 用于定义枚举类型日志级别 |
| Fields      | map[string]any | 分级日志需要处理的数据   |

#### 枚举类型

LoggerLevel

| 标识符     | 值   | 说明    |
| ---------- | ---- | ------- |
| LevelDebug | 0    | DEBUG级 |
| LevelInfo  | 1    | INFO级  |
| LevelError | 2    | ERROR级 |

#### 接口

LoggerFormatter

分级日志格式化器的类型：暂时有TextFormatter（默认）和JsonFormatter

| 方法                              | 说明                         |
| --------------------------------- | ---------------------------- |
| Format(*LoggerFormatParam) string | 日志的数据处理（输出格式化） |

#### 数据结构

LoggerFormatParam

| 属性         | 类型        | 说明                   |
| ------------ | ----------- | ---------------------- |
| Message      | any         | 日志信息               |
| IsColor      | bool        | 是否需要显示颜色       |
| Level        | LoggerLevel | 日志级别               |
| LoggerFields | Fields      | 分级日志需要处理的数据 |

Logger

| 属性         | 类型            | 说明                     |
| ------------ | --------------- | ------------------------ |
| Formatter    | LoggerFormatter | 分级日志格式化器         |
| Level        | LoggerLevel     | 日志级别                 |
| Outs         | []*LoggerWriter | 日志输出的写入器数组     |
| LoggerFields | Fields          | 分级日志需要带有属性数据 |
| logPath      | string          | 日志本地打印地址         |
| LogFileSize  | int64           | 单个日志文件最大的大小   |

LoggerWriter

| 属性  | 类型        | 说明     |
| ----- | ----------- | -------- |
| Level | LoggerLevel | 日志级别 |
| Out   | io.Writer   | 写入器   |

#### 方法/函数

`func New() *Logger` 
返回一个创建空的分级日志格式化器；

`func Default() *Logger`
返回默认的分级日志格式化器；
默认等级为DEBUG，写入器Outs中加入等级为DEBUG的os.Stout的（输出）写入器；
默认的日志格式化形式为TextFormatter；

`func (l *Logger) WithFields(fields Fields) *Logger`
返回一个带有处理的数据的日志格式化器，原有创建的日志格式化器参数都会赋到新的日志格式化器上；

`func (l *Logger) Print(level LoggerLevel, msg any)`
日志的格式化输出；当l.Level大于level时将不打印对应日志；msg是日志需要附加的打印信息；
根据写入器的类型以及日志级别会自动判断是打印到控制台还是打印到本地，会分级打印日志；

`func (l *Logger) Debug(msg any)`
DEBUG日志的输出，内部调用Print(LevelDebug)；

`func (l *Logger) Info(msg any)`
INFO日志的输出，内部调用Print(LevelInfo)；

`func (l *Logger) Error(msg any)`
ERROR日志的输出，内部调用Print(LevelError)；

`func FileWriter(name string) io.Writer`
返回一个向目标（name，默认和项目同级）地址打印输出日志文件name.log的文件写入器；文件权限 0644 ；

`func (l *Logger) SetLogPath(logPath string)`
打印四种日志；日志打印地址logPath会存入到Logger中，并且加入四种日志写入器；
①所有日志：log.log
②DEBUG级日志：debug.log
③INFO级日志：info.log
④ERROR级日志：error.log

`func (l *Logger) CheckFileSize(w *LoggerWriter)`
检查文件大小是否过大，过大就将对应的写入进行更新以达到日志分割的目的，新的日志文件将加上时间戳；（以引用传递的方式改变了l.Outs内部对应的写入器）

`func (l LoggerLevel) Level() string`
返回当前的日志级别；

`func (f *LoggerFormatParam) DefaultLevelColor() string`
返回当前级别默认的级别标识颜色；

`func (f *LoggerFormatParam) DefaultMsgColor() string`
返回当前级别默认的信息颜色；

### json.go

#### 数据结构

JsonFormatter

| 属性        | 类型 | 说明                       |
| ----------- | ---- | -------------------------- |
| TimeDisplay | bool | 是否在JSON中加入时间的显示 |

#### 方法/函数

`func (f *JsonFormatter) Format(param *LoggerFormatParam) string`
实现LoggerFormatter的Format方法，对日志输出进行JSON格式化；

### text.go

#### 数据结构

TextFormatte \[暂] 为空

#### 方法/函数

`func (f *TextFormatter) Format(param *LoggerFormatParam) string`
实现LoggerFormatter的Format方法，对日志输出进行普通的文本格式化输出；

------

## orm

### condition.go

方法/函数

`func (s *SpxSession) writeConditionParam(field string, value any) *SpxSession`
[代码复用]写入使用条件时的占位符以及对应的值

`func (s *SpxSession) Where(field string, value any) *SpxSession`
where条件

`func (s *SpxSession) And(field string, value any) *SpxSession`
and条件

`func (s *SpxSession) Or(field string, value any) *SpxSession`
or条件

`func (s *SpxSession) Like(field string, value any) *SpxSession`
like模糊查询 %value%

`func (s *SpxSession) LikeRight(field string, value any) *SpxSession`
like右模糊查询 value%

`func (s *SpxSession) LikeLeft(field string, value any) *SpxSession`
like左模糊查询 %value

`func (s *SpxSession) Group(field ...string) *SpxSession`
group分组

`func (s *SpxSession) OrderDesc(field ...string) *SpxSession`
降序排列

`func (s *SpxSession) OrderAsc(field ...string) *SpxSession`
升序排列

`func (s *SpxSession) Order(field ...string) *SpxSession`
降序和升序结合；[规范] (field1,desc,field2,asc) desc 必须在 asc前面

### delete.go

方法/函数

`func (s *SpxSession) Delete()`
删除操作(delete from table_name [where 需要调用])

### function.go

方法/函数

`func (s *SpxSession) Aggregate(funcName, field string) (int64, error)`
统计函数，下面六种统计函数名填入funcName可以实现对应操作；

`func (s *SpxSession) Average(field string) (int64, error)`
返回某列的平均值

`func (s *SpxSession) Count(field string) (int64, error)`
返回某列的行数

`func (s *SpxSession) Max(field string) (int64, error)`
返回某列的最大值

`func (s *SpxSession) Min(field string) (int64, error)`
返回某列的最小值

`func (s *SpxSession) Sum(field string) (int64, error)`
返回某列值之和

### insert.go

方法/函数

`func (s *SpxSession) Insert(data any) (int64, int64, error)`
插入单条数据，返回id、影响行数、err；

`func (s *SpxSession) InsertBatch(data []any) (int64, int64, error)`
批量插入数据，返回id、影响行数、err；

### native.go

提供原生 SQL 支持

方法/函数

`func (s *SpxSession) Exec(sql string, values ...any) (int64, error)`
增删改的原生sql支持；

`func (s *SpxSession) QueryRow(sql string, data any, queryValues ...any) error`
查询一行的原生sql支持；

### orm.go

#### 数据结构

SpxDb

| 属性   | 类型           | 说明       |
| ------ | -------------- | ---------- |
| db     | *sql.DB        | 数据库句柄 |
| logger | *sqlLog.Logger | 分级日志   |

SpxSession

使每次对数据库的操作都在一个会话内完成 相互独立

| 属性            | 类型            | 说明                     |
| --------------- | --------------- | ------------------------ |
| db              | *SpxDb          | 带有数据库句柄的数据结构 |
| tx              | *sql.Tx         | 数据库事务               |
| beginTx         | bool            | 是否开启了数据库事务     |
| tableName       | string          | 数据库表名               |
| fieldName       | []string        | 数据表字段名             |
| placeHolder     | []string        | 占位符(插入操作时)       |
| values          | []any           | 占位符所需的值           |
| updateParam     | strings.Builder | 更新参数                 |
| conditionParam  | strings.Builder | 条件参数                 |
| conditionValues | []any           | 条件占位符所需的值       |

#### 方法/函数

`func Open(driverName string, source string) *SpxDb`
打开并连接数据库、初始化默认分级日志并返回

`func (db *SpxDb) Close() error`
关闭数据库连接

`func (db *SpxDb) SetMaxIdleConns(n int)`
设置最大空闲连接数，默认不配置，2两个最大空闲连接

`func (db *SpxDb) SetMaxOpenConns(n int)`
设置最大连接数，默认不设置，不限制最大连接数

`func (db *SpxDb) SetConnMaxLifetime(duration time.Duration)`
设置连接最大存活时间

`func (db *SpxDb) SetConnMaxIdleTime(duration time.Duration)`
设置空闲连接最大存活时间

`func (db *SpxDb) New(table string) *SpxSession`
返回一个新的数据库会话，可以附带有表名；

`func (s *SpxSession) Table(name string) *SpxSession`
[废弃]可在New()操作之后进行表的选择；

### select.go

方法/函数

`func (s *SpxSession) SelectOne(data any, fields ...string) error`
查询单行；data为对应数据结构，fields为需要查询的字段

`func (s *SpxSession) Select(data any, fields ...string) ([]any, error)`
查询多行

### transaction.go

提供事务支持

方法/函数

`func (s *SpxSession) Begin() error`
事务开启

`func (s *SpxSession) Commit() error`
事务提交

`func (s *SpxSession) Rollback() error`
事务回滚

### update.go

方法/函数

`func (s *SpxSession) Update(data ...any) (int64, int64, error)`
更新操作；
支持两种模式：1.一个字段，两参数：字段 + 值； 2.两个字段及以上直接传入相应数据结构；

### utils.go

#### 枚举常量

用于复用代码时不同类型的选择()

| 标识符      | 说明       |
| ----------- | ---------- |
| Insert      | 插入型     |
| InsertBatch | 批量插入型 |
| Update      | 更新型     |
| Delete      | 删除型     |
| Select      | 查询型     |
| SelectOne   | 单个查询型 |

#### 方法/函数

`func IsAutoId(id any) bool`
判断ID字段是否自动增长；

`func Name(name string) string`
在数据结构没有使用Tag时对属性名进行转换：驼峰命名->下划线命名；

`func (s *SpxSession) execute(option int, query string) (int64, int64, error)`
[代码复用]对增删改的代码进行复用，使代码简洁，内部流程为：创建准备好的语句->执行语句->最后一个ID(只在插入时有效)->影响行数
返回 id 行数 err

`func (s *SpxSession) filling(option int, data any)`
填充占位符；

`func (s *SpxSession) batchFilling(data []any)`
批量操作的占位符填充；

`func (s *SpxSession) selectComponent(data any, fields ...string) (*sql.Rows, []string, error)`
[代码复用]查询代码的复用，内部流程为：创建准备好的语句->执行准备好的查询语句->得到列名；
返回 查询结果 列名 err

`func (s *SpxSession) convertType(i, j int, values []any, tVal reflect.Type) reflect.Value`
类型转换；用在查询中，为了使查询结果能够绑定到数据结构上，需要利用反射进行类型转换；

`func (s *SpxSession) nextFill(option int, data any, t reflect.Type, rows *sql.Rows, cols []string) (result any, err error)`
查询下一行并返回需要绑定的数据,一次得到一行的查询结果；用于rows.Next()的(循环)判断；



------

## pool

### pool.go

#### 常量

| 常量名        | 常量值 | 说明                           |
| ------------- | ------ | ------------------------------ |
| DefaultExpire | 1      | worker默认过期失效时间，单位 s |

#### 全局变量

| 变量名             | 变量值                                | 说明                        |
| ------------------ | ------------------------------------- | --------------------------- |
| ErrorInValidCap    | errors.New("cap can not < 0")         | 协程池的容量不能小于0       |
| ErrorInValidExpire | errors.New("expire time can not < 0") | worker过期失效时间不能小于0 |
| ErrorPoolClosed    | errors.New("pool has been released")  | 协程池已经被释放了          |
| ErrorRestart       | errors.New("pool restart error")      | 协程池重启错误              |

#### 数据结构

sign

用作标记协程池是否被释放，没有属性；

Pool

| 属性         | 类型          | 说明                                                         |
| ------------ | ------------- | ------------------------------------------------------------ |
| workers      | []*Worker     | 空闲worker                                                   |
| workerCache  | sync.Pool     | worker 缓存，提高性能                                        |
| cap          | int32         | pool worker 的最大容量                                       |
| running      | int32         | 正在工作的 worker 数量                                       |
| expire       | time.Duration | 过期时间 空闲 worker 超过该时间就回收                        |
| release      | chan sign     | g释放资源 pool 就不能使用了                                  |
| lock         | sync.Mutex    | 保护 pool 中相关资源的安全                                   |
| once         | sync.Once     | 保证 release 只能调用一次，不能多次调用                      |
| cond         | *sync.Cond    | 基于互斥锁/读写锁实现的条件变量,协调想要访问共享资源的那些 goroutine |
| PanicHandler | func()        | panic处理器函数，可供用户自定义                              |

#### 方法/函数

`func NewPool(cap int) (*Pool, error)`
创建一个默认的协程池，容量自定，过期失效时间为默认的时间；

`func NewTimePool(cap int, expire int) (*Pool, error)`
创建一个协程池，需提供容量和过期时间；

`func (p *Pool) expireWorker()`
定期清理长时间空闲的worker，防止大量 worker 的占用，造成性能的浪费；若协程池被释放，则该协程停止工作；

`func (p *Pool) createWorker() *Worker`
提供了代码的复用，在没有空闲的worker且running < cap时创建一个新的worker；

`func (p *Pool) GetWorker() *Worker`
返回一个空闲的worker来执行任务；
①有空闲的 worker 直接获取；
②没有空闲 worker 新建一个 worker；
③如果正在运行的 worker ==  cap ，阻塞等待 worker 释放；

`func (p *Pool) waitIdleWorker() *Worker`
如果正在运行的 worker ==  cap ，阻塞等待 worker 释放；使用sync.Cond；

`func (p *Pool) Submit(task func()) error`
提交任务给协程，让协程池中的worker执行任务；

`func (p *Pool) incRunning()`
原子操作，正在工作的worker+1；

`func (p *Pool) decRunning()`
原子操作，正在工作的worker-1；

`func (p *Pool) PutWorker(w *Worker)`
将工作完成的worker重新放入到协程池中，并通知有空闲的worker可以执行任务；

`func (p *Pool) Release()`
释放协程池；

`func (p *Pool) IsReleased() bool`
协程池是否被释放；

`func (p *Pool) Restart() error`
重启协程池；

`func (p *Pool) Running() int`
返回当前正在工作的worker；

`func (p *Pool) Free() int`
返回当前空闲的worker位置（可创建）；

### worker.go

#### 数据结构

| 属性     | 类型        | 说明               |
| -------- | ----------- | ------------------ |
| pool     | *Pool       | 绑定对应的协程池   |
| task     | chan func() | 任务队列           |
| lastTime | time.Time   | 最后执行任务的时间 |

#### 方法/函数

`func (w *Worker) run()`
让worker开始执行任务；

`func (w *Worker) running()`
执行任务，具体的任务执行过程；

------

## render

### render.go

#### 接口

Render

实现Render接口后可以执行相应的渲染方法。

| 方法                                                | 说明                                   |
| --------------------------------------------------- | -------------------------------------- |
| Render(w http.ResponseWriter, statusCode int) error | 渲染器，执行后响应请求将数据渲染到页面 |
| WriteContentType(w http.ResponseWriter)             | 写入响应时的Content-Type               |

#### 方法/函数

`func writeContentType(w http.ResponseWriter, value string)`
写入响应时的Content-Type，依据value值变化；render下都可以调用该函数；

### redirect.go

#### 数据结构

Redirect

| 属性     | 类型          | 说明       |
| -------- | ------------- | ---------- |
| Request  | *http.Request | HTTP请求   |
| Location | string        | 重定向地址 |

#### 方法/函数

`func (r *Redirect) Render(w http.ResponseWriter, statusCode int) error`
① (statusCode < 300 || statusCode > 380) && statusCode != 20 ：在这些状态码情况下重定向失败；
② 重定向到Location地址

`func (r *Redirect) WriteContentType(w http.ResponseWriter)`
重定向不需要具体实现该方法，方法体为空

### html.go

#### 数据结构

HTML

| 属性       | 类型               | 说明                                                         |
| ---------- | ------------------ | ------------------------------------------------------------ |
| Name       | string             | 被渲染的html文件名                                           |
| Data       | any                | 需要被渲染的数据                                             |
| Template   | *template.Template | 导入的本地html模板文件<br />（Context.engine.HTMLRender.Template <- Engine.HTMLRender=template.Template） |
| IsTemplate | bool               | 是否使用模板(Context.HTML和Context.Template的区分)           |

#### 方法/函数

`func (h *HTML) Render(w http.ResponseWriter, statusCode int) error`
①h.IsTemplate==true：调用tempate(h).Template.ExecuteTemplate()关联模板进行html渲染
②h.IsTemplate==false：将传入的数据通过字符串解析后用http.ResponseWriter.Write()直接进行渲染

`func (h *HTML) WriteContentType(w http.ResponseWriter)`
写入Content-Type：text/html; charset=utf-8

### json.go

#### 数据结构

JSON

| 属性 | 类型 | 说明     |
| ---- | ---- | -------- |
| Data | any  | JSON数据 |

#### 方法/函数

`func (j *JSON) Render(w http.ResponseWriter, statusCode int) error`
调用json.Marshal()将数据编码转[]byte后进行渲染

`func (j *JSON) WriteContentType(w http.ResponseWriter) `
写入Content-Type：application/json; charset=utf-8

### string.go

#### 数据结构

String

| 属性   | 类型   | 说明                                   |
| ------ | ------ | -------------------------------------- |
| Format | string | 格式本文（带有%s %d %v等格式化字符串） |
| Data   | []any  | 格式化字符串值                         |

#### 方法/函数

`func (s *String) Render(w http.ResponseWriter, statusCode int) error `
①len(Data)>0说明需要进行格式化，所以进行格式化String渲染(Fprintf)
②否则说明不需要格式化，Write即可

`func (s *String) WriteContentType(w http.ResponseWriter)`
写入Content-Type：text/plain; charset=utf-8

### xml.go

#### 数据结构

| 属性 | 类型 | 说明    |
| ---- | ---- | ------- |
| Data | any  | XML数据 |

#### 方法/函数

`func (x *XML) Render(w http.ResponseWriter, statusCode int) error`
使用xml.NewEncoder(w).Encode(x.Data)将传入数据xml格式化渲染；NewEncoder(w)相当于传入了Write()；

`func (x *XML) WriteContentType(w http.ResponseWriter)`
写入Content-Type：application/xml; charset=utf-8

------

## serror

本框架自定义错误

### errors.go

#### 自定义类型

| 类型名    | 具体类型                 | 说明               |
| --------- | ------------------------ | ------------------ |
| ErrorFunc | func(spxError *SpxError) | 自定义错误处理方法 |

#### 数据结构

SpxError

| 属性      | 类型      | 说明               |
| --------- | --------- | ------------------ |
| err       | error     | 自定义错误         |
| ErrorFunc | ErrorFunc | 自定义错误处理方法 |

#### 方法/函数

`func Default() *SpxError`
默认返回一个空的自定义错误结构；

`func (e *SpxError) Error() string`
返回自定义错误的错误信息；

`func (e *SpxError) Put(err error)`
将自定义错误放入，若err不空，则放入e中并造成panic；

`func (e *SpxError) check(err error)`
若err不空，则放入e中并造成panic；

`func (e *SpxError) Result(ef ErrorFunc)`
暴露该方法让用户能够自己定义 自定义错误 的处理方式；

`func (e *SpxError) ExecuteResult()`
调用自定义错误方法去处理自定义错误；

------

## token

### token.go

调用了github.com/golang-jwt/jwt/v4

#### 常量

| 常量名   | 常量值        | 说明         |
| -------- | ------------- | ------------ |
| JWTToken | "spxgo_token" | j.CookieName |

#### 数据结构

JwtHandler

| 属性           | 类型                                           | 说明                                                         |
| -------------- | ---------------------------------------------- | ------------------------------------------------------------ |
| Algorithm      | string                                         | jwt使用的算法                                                |
| TimeOut        | time.Duration                                  | jwt过期时间                                                  |
| TimeFunc       | func() time.Time                               | jwt调用后返回创建时间                                        |
| RefreshTimeOut | time.Duration                                  | jwt刷新过期时间，即在没超过此时间之前可以对原有token进行刷新 |
| RefreshKey     | string                                         | 刷新的键，用于获取jwt令牌                                    |
| Key            | []byte                                         | 算法加密值                                                   |
| PrivateKey     | string                                         | 算法加密私钥                                                 |
| SendCookie     | bool                                           | 是否生成cookie                                               |
| Authenticator  | func(c *spxgo.Context) (map[string]any, error) | jwt令牌中需要传输的数据                                      |
| CookieName     | string                                         | Cookie的名称，Cookie一旦创建，名称便不可更改                 |
| CookieMaxAge   | int                                            | Cookie失效的时间                                             |
| CookieDomain   | string                                         | 可以访问该Cookie的域名                                       |
| SecureCookie   | bool                                           | 该Cookie是否仅被使用安全协议传输                             |
| CookieHTTPOnly | bool                                           | cookie是否可通过客户端脚本访问                               |
| Header         | string                                         | 需要获得的请求头内容                                         |
| AuthHandler    | func(c *spxgo.Context, err error)              | 请求失败时可供选择的处理方法                                 |

#### 方法/函数

`func (j *JwtHandler) LoginHandler(c *spxgo.Context) (*JwtResponse, error)`
为登陆请求提供JWT令牌；

`func (j *JwtHandler) usingPublicKeyAlgo() bool`
判断算法是否要使用私钥；

`func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error)`
返回刷新jwt令牌；

`func (j *JwtHandler) LogoutHandler(c *spxgo.Context)`
登出处理，使cookie过期；

`func (j *JwtHandler) RefreshHandler(c *spxgo.Context) (*JwtResponse, error)`
刷新jwt令牌的过期时间等信息；

`func (j *JwtHandler) AuthInterceptor(next spxgo.HandlerFunc) spxgo.HandlerFunc`
JWT中间件，验证请求是否认证；

------

## auth.go

### 数据结构

Accounts

| 属性          | 类型              | 说明                               |
| ------------- | ----------------- | ---------------------------------- |
| UnAuthHandler | func(c *Context)  | 提供给用户自定义未认证时的失败情况 |
| Users         | map[string]string | 保存用户的 用户名：密码            |

### 方法/函数

`func (a *Accounts) BasicAuth(next HandlerFunc) HandlerFunc`
Basic认证中间件

`func (a *Accounts) unAuthHandler(c *Context)`
默认的未认证处理；

`func BasicAuth(username, password string) string`
返回base64加密后的Basic认证用户名和密码；

------

## context.go

### 常量

| 常量命                 | 常量值          | 说明         |
| ---------------------- | --------------- | ------------ |
| defaultMaxMemory       | 32 << 20 // 32M | 默认分配内存 |
| defaultMultipartMemory | 32 << 20 // 32M | 默认分配内存 |

### 数据结构

Context

| 属性                  | 类型                | 说明                                                         |
| --------------------- | ------------------- | ------------------------------------------------------------ |
| W                     | http.ResponseWriter | ResponseWriter接口被HTTP处理器用于构造HTTP回复               |
| R                     | *http.Request       | Request类型代表一个服务端接受到的或者客户端发送出去的HTTP请求 |
| engine                | *Engine             | 框架引擎， 通过sync.Pool.New使Context获取了框架引擎          |
| queryCache            | url.Values          | GET请求查询的参数和表单的属性，go提供了原生query参数的map支持 |
| postFormCache         | url.Values          | POST请求传输的数据                                           |
| StatusCode            | int                 | 返回的HTTP状态码                                             |
| DisallowUnknownFields | bool                | 是否开启处理参数中有而结构体中没有属性                       |
| IsValidate            | bool                | 是否开启处理结构体中必要的属性（通过在属性后使用 \`spxgo:"required"\` 这个 tag 开启） |
| Logger                | *spxLog.Logger      | 分级日志格式化器                                             |
| Keys                  | map[string]any      | 认证信息                                                     |
| mutex                 | sync.RWMutex        | 用读写锁来保证Key的读写                                      |

### 方法/函数

`func (c *Context) Set(key string, value any)`
将Basic信息加入到Context.Keys中；

`func (c *Context) Get(key string) (value any, ok bool)`
返回键为key的Basic信息；

`func (c *Context) SetBasicAuth(username, password string) `
在请求头上加上Basic认证；

`func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)`
生成cookie，本质上是调用了http.SetCookie()；

`func (c *Context) Render(statusCode int, r render.Render) error`
调用对应类型的render.Render渲染器并传递状态码；
目前有render.HTML、render.String、render.JSON、render.XML、render.Redirect五种类型;

`func (c *Context) HTML(status int, data string) error`
调用HTML渲染器实现对HTML的渲染，不使用template模板（text/html）；

`~~func (c *Context) HTMLTemplateFiles(name string, data any, filenames ...string) error~~`
~~\[废弃]以template.Tempalte.ParseFiles()的方式依据本地html模板渲染HTML，缺点是要一个个写本地文件名~~

`func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) error `
~~\[废弃]以template.Tempalte.ParseGlob()的方式依据pattern自动匹配本地html模板文件，可以一次性导入多个模板~~

`func (c *Context) Template(status int, name string, data any) error`
调用HTML渲染器实现对HTML的渲染，使用template模板（text/html，通过Context上下文中的engine传入本地html模板）；

`func (c *Context) JSON(status int, data any) error`
调用JSON渲染器实现对JSON的渲染（application/json）；

`func (c *Context) XML(status int, data any) error`
调用XML渲染器实现对XML的渲染（application/xml）；

`func (c *Context) Redirect(status int, location string) error`
调用重定向渲染器实现重定向；

`func (c *Context) StringOld(status int, format string, values ...any) error`
~~\[废弃]String渲染代码重构前的旧代码~~

`func (c *Context) String(status int, format string, values ...any) error`
调用String渲染器实现对String的渲染（text/plain）；

`func (c *Context) File(filename string)`
调用http.ServeFile()回复请求filename指定的文件或者目录的内容，不支持自定义文件名；

`func (c *Context) FileAttachment(filepath, filename string)`
调用http.ServeFile()回复请求filename指定的文件或者目录的内容，支持自定义文件名filename；\[注意编码的转换]

`func (c *Context) FileFromFS(filepath string, fs http.FileSystem)`
相对系统文件路径进行渲染；fs 一般使用http.Dir()打开的操作系统接口提供文件访问服务；

`func (c *Context) initQueryCache()`
 初始化Context.queryCache；将GET查询的参数和表单的属性储到上下文；若空则初始化为url.Values{}

`func (c *Context) GetQuery(key string) string`
返回Query中key关联的第一个值，若为空则返回一个空串；

`func (c *Context) QueryArray(key string) []string`
返回Query中key关联的所有值；

`func (c *Context) GetQueryArray(key string) ([]string, bool)`
返回Query中key关联的所有值和queryCahche中是否存在key的bool值；

`func (c *Context) GetDefaultQuery(key, defaultValue string) string`
返回Query中key关联的第一个值，若key不存在则返回defaultValue；

`func (c *Context) get(m map[string][]string, key string) (map[string]string, bool)`
将参数解析成map并返回，bool为map是否存在，由于字典形式为[string]string所以只能获取关联的第一个值；
其中key为参数中map的名字，如usr[id]=1中的usr，

`func (c *Context) GetQueryMap(key string) (map[string]string, bool)`
提供更为复杂的Query的获取，形如.../?usr[id]=1&usr[id]=2&usr[name]=spxzx...；其中key为usr；
返回解析的Query map、map是否存在的bool值，由于字典形式为[string]string所以只能获取关联的第一个值；

`func (c *Context) QueryMap(key string) map[string]string`
和GetQueryMap一致但是不返回bool值；

`func (c *Context) initPostFormCache()`
初始化Context.postFormCache；将POST查询的参数和表单的属性储到上下文；若空则初始化为url.Values{}

`func (c *Context) GetPostForm(key string) (string, bool)`
返回postFormCache(params)中key关联的第一个值及bool值；

`func (c *Context) PostFormArray(key string) []string`
返回params中key关联的所有值；

`func (c *Context) GetPostFormArray(key string) ([]string, bool)`
返回params中key关联的所有值及bool值；

`func (c *Context) PostFormMap(key string) map[string]string`
返回解析params中名为key的map，由于字典形式为[string]string所以只能获取关联的第一个值；

`func (c *Context) GetPostFormMap(key string) (map[string]string, bool)`
返回解析params中名为key的map、map是否存在的bool值，由于字典形式为[string]string所以只能获取关联的第一个值；

`func (c *Context) MultipartFormFiles() (*multipart.Form, error)`
返回解析后的多部分表单，包括文件上传。

`func (c *Context) FormFile(key string) *multipart.FileHeader`
返回提供的表单键（key）的第一个文件；FileHeader 描述了多部分请求的文件部分；

`func (c *Context) FormFiles(key string) []*multipart.FileHeader`
返回提供的表单键（key）的所有文件；若没有文件则返回[]*multipart.FileHeader{}；

`func (c *Context) SaveUploadFile(file *multipart.FileHeader, dst string) error`
单个文件保存的实现；将表单单个文件保存到本地项目dst目录下；

`func (c *Context) SaveAllUploadFiles(files []*multipart.FileHeader, dst string) error`
将表单所有文件保存到本地项目dst目录下；

`func (c *Context) MustBindWith(obj any, bind binding.Binding) error`
绑定绑定器，验证绑定器是否绑定成功；

`func (c *Context) ShouldBindWith(obj any, bind binding.Binding) error`
绑定对应的绑定器，绑定器用于解析query/params后将其以对应形式绑定到数据结构中；

`func (c *Context) BindJson(obj any) error`
绑定JSON绑定器；一定要传入指针/引用类型；

`func (c *Context) BindXML(obj any) error`
绑定XML绑定器；一定要传入指针/引用类型；

`func (c *Context) Fail(statusCode int, msg string)`
用于错误/失败响应的String渲染；一般传入的statusCode都和错误有关，信息也是错误信息；

## log.go

日志中间件

### 常量

| 常量名    | 常量值                           |
| --------- | -------------------------------- |
| greenBg   | 一系列颜色值，带Bg的改变背景颜色 |
| whiteBg   |                                  |
| yellowBg  |                                  |
| redBg     |                                  |
| blueBg    |                                  |
| magentaBg |                                  |
| cyanBg    |                                  |
| green     |                                  |
| white     |                                  |
| yellow    |                                  |
| red       |                                  |
| blue      |                                  |
| magenta   |                                  |
| cyan      |                                  |
| reset     | 颜色清空                         |

### 全局变量

| 变量名              | 变量类型                                | 变量值                                   | 说明               |
| ------------------- | --------------------------------------- | ---------------------------------------- | ------------------ |
| defaultWriter       | io.Writer                               | os.Stdout                                | 默认写入器         |
| defaultLogFormatter | func(params *LogFormatterParams) string | [defaultLogFormatter](#defaultFormatter) | 默认的日志格式化器 |

### 自定义类型

| 类型名          | 具体类型                                | 说明         |
| --------------- | --------------------------------------- | ------------ |
| LoggerFormatter | func(params *LogFormatterParams) string | 日志格式化器 |

### 数据结构

LoggingConfig

| 属性名    | 类型            | 说明         |
| --------- | --------------- | ------------ |
| Formatter | LoggerFormatter | 日志格式化器 |
| out       | io.Writer       | 输出写入器   |

LogFormatterParams

| 属性名         | 类型          | 说明         |
| -------------- | ------------- | ------------ |
| Request        | *http.Request | 请求         |
| TimeStamp      | time.Time     | 时间戳       |
| StatusCode     | int           | 响应状态码   |
| Latency        | time.Duration | 响应时间     |
| ClientIP       | net.IP        | 客户端IP     |
| Method         | string        | RESTful方式  |
| Path           | string        | 路由路径     |
| IsDisplayColor | bool          | 是否显示颜色 |

### 方法/函数

`func (p *LogFormatterParams) StatusCodeColor() string`
根据状态码设置日志打印的状态码颜色；

`func (p *LogFormatterParams) ResetColor() string`
返回清空颜色的颜色码；

<a name="defaultFormatter">`defaultLogFormatter = func(params *LogFormatterParams) string`</a>
默认日志格式化输出器，可以对日志进行格式化输出；

`func LoggingWithConfig(conf LoggingConfig, next HandlerFunc) HandlerFunc`
配置日志的各种参数，默认的Formatter为defaultLogFormatter，，默认的out写入器为io.Writer；
fmt.Fprintln(out, formatter(param))输出日志；

`func Logging(next HandlerFunc) HandlerFunc`
返回默认的日志格式化处理器；可以直接注册中间件使用；

## recovery.go

错误恢复中间件

### 方法/函数

`func detailMsg(err any) string`
调用 goroutine 堆栈得到栈帧信息，对其处理后可以得到panic的错误详细的代码地址；

`func Recovery(next HandlerFunc) HandlerFunc`
捕获panic错误并进行恢复和日志输出，除了go本身的错误还可以捕获自定义的panic错误；
①go本身错误会使用Error级日志进行输出并且调用Context.Fail(500，msg)进行渲染；
②自定义错误会根据用户自定义的处理方法去进行处理；

## spx.go

### 常量

| 常量命    | 常量值 | 说明                 |
| --------- | ------ | -------------------- |
| MethodAny | "ANY"  | RESTful 方法可为任意 |

### 数据结构

Engine

| 属性         | 类型              | 说明                                                         |
| ------------ | ----------------- | ------------------------------------------------------------ |
| _            | [router](#router) | 无属性名，路由                                               |
| funcMap      | template.FuncMap  | 存储template.FuncMap映射                                     |
| HTMLRender   | renderHTML        | html渲染器                                                   |
| pool         | sync.Pool         | sync.Pool 用于存储那些被分配了但是还没有被使用，<br />但是未来可能使用的值，这样可以不用再次分配内存，提高效率 |
| Logger       | *spxLog.Logger    | 分级日志器                                                   |
| middles      | []MiddlewareFunc  | 默认组通用中间件                                             |
| errorHandler | ErrorHandler      | 错误处理器                                                   |

<a name="router">router</a>

| 属性         | 类型                           | 说明                                    |
| ------------ | ------------------------------ | --------------------------------------- |
| routerGroups | [[]*routerGroup](#routerGroup) | 路由组                                  |
| engine       | *Engine                        | 框架引擎，\[暂]用于默认引擎中调用中间件 |

<a name="routerGroup">routerGroup</a>

| 属性               | 类型                                                        | 说明                                                         |
| ------------------ | ----------------------------------------------------------- | ------------------------------------------------------------ |
| name               | string                                                      | 路由组名"/name(/......)"                                     |
| treeNode           | *treeNode                                                   | 前缀树，用来构建路由路径，便于处理                           |
| handlerFuncMap     | map[string]map\[string\][HandlerFunc](#HandlerFunc)         | { "name": { "method": HandleFunc } } <br />路由、请求方法、处理器之间的映射，用于判断请求方法 |
| middlewaresFuncMap | map[string]map\[string][\][MiddlewareFunc](#MiddlewareFunc) | { "name": { "method": []MiddlewareFunc } }<br />路由、请求方法、中间件处理器之间的映射，用于组路由级别中间件 |
| middlewares        | []MiddlewareFunc                                            | 组通用中间件                                                 |

### 自定义类型

| 类型名                                      | 具体类型                                  | 说明                                                         |
| ------------------------------------------- | ----------------------------------------- | ------------------------------------------------------------ |
| <a name="HandlerFunc">HandlerFunc</a>       | func(c *Context)                          | 处理器函数                                                   |
| <a name="MiddlewareFunc">MiddlewareFunc</a> | func(handlerFunc HandlerFunc) HandlerFunc | 中间件处理器函数， 传入 HandlerFunc 处理完后再将其返回 ->达成影响代码 |

### 方法/函数

`func (r *router) Group(name string) *routerGroup`
创建一个路由组并返回；name 是路由组的名字，创建的路由组会将默认的中间件(Logger、Recovery)加入，将该组加入router.routerGroups中

`func (r *routerGroup) handle(name string, method string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc)`
处理请求方法和中间件；router.handlerFuncMap[name]和router.middlewaresFuncMap[name]为空则make分配内存并初始化，如果已经存在该路由则panic，将传入的处理器方法和（多个）中间件进行存储，并将该路由名以前缀树方式存储；

`func (r *routerGroup) Use(middlewareFunc ...MiddlewareFunc)`
将传入的（组通用）中间件全部加入到路由组；

`func (r *routerGroup) methodHandle(name string, method string, handleFunc HandlerFunc, c *Context)`
处理（匹配）组通用级中间件和组组路由级中间件，最后执行使用中间件处理完毕后的处理器方法，提供服务; 
\[handle = func(前置处理器->handle(c)<-后置处理器); 最后执行handle(c)\] \*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*\*

`func (r *routerGroup) Any(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc)`
路由使用该请求方式时 RESTful 各种请求方法都可以通过; handlerFunc 为处理器具体处理函数, middlewareFunc是中间件;
除了Any还有Get()、Post()、Delete()、Put()、Patch()、Options()、Head()请求方法，省略。

`func New() *Engine`
创建一个框架引擎并返回；

`func (e *Engine) allocateContext() any`
返回Context上下文；上下文中的engine属性赋值为当前的engine e，该方法用于func New() *Engine中engine.pool.New的初始化，[以防止Context在程序中频繁被创建]

`func (e *Engine) Use(middlewareFunc ...MiddlewareFunc)`
往 e 中加入要默认使用的中间件;

`func Default() *Engine`
创建本框架默认的引擎；默认引擎中带有分级、普通日志功能，\[engine.router.engine = engine] 使能够调用默认的中间件；

`func (e *Engine) SetFuncMap(funcMap template.FuncMap) `
~~\[废弃]设置 template.FuncMap map[string]any 映射键值对~~

`func (e *Engine) SetHTMLRender(t *template.Template)`
设置HTML渲染模板；

`func (e *Engine) LoadTemplate(pattern string)`
加载需要的HTML模板到engine中，方便全局调用；解析匹配pattern的文件里的模板定义（本地html模板）；

`func (e *Engine) httpRequestHandle(c *Context, w http.ResponseWriter, r *http.Request)`
**[Tips]该方法目前可能还存在一个极端的路由匹配问题；**具体处理请求，匹配ANY,GET,POST等请求方法，匹配中间件；
①路由资源不存在，=> 路由地址 not found；
②路由请求所用的方法不支持 => #{method} not allowed；
③路由资源存在，经过methodHandle方法后完成请求的处理 ；

`func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request)`
实现http.server下Handler的ServeHTTP()方法，每次监听到请求时都会进入这个方法中进行处理；
具体实现中上下文赋值为e.pool.Get().(*Context)\[最后e.pool.Put(c),将上下文放入pool中]，以防止Context在程序中频繁被创建，其他有关信息也都被传入到上下文中，之后调用httpRequestHandle()进行请求的处理；

`func (e *Engine) Run()`
开启框架支持，注册HTTP处理器handler和对应的模式pattern，监听TCP地址addr；

`func (e *Engine) RunTLS(addr, certFile, keyFile string)`
同上，但支持的是https;

## tree.go

### 数据结构

treeNode

| 属性       | 类型        | 说明                 |
| ---------- | ----------- | -------------------- |
| name       | string      | 当前结点路径名       |
| routerName | string      | 当前结点累积的路径名 |
| isLeaf     | bool        | 当前结点是否叶子结点 |
| children   | []*treeNode | 孩子结点             |

### 方法/函数

`func (t *treeNode) Put(path string)`
将当前的路径存储到前缀树中；

`func (t *treeNode) Get(path string) *treeNode`
获取该路径对应的叶结点；

## utils.go

### 工具函数

`func subStringLast(str string, substr string) string`
用来处理路由组路径，即将路由组名与其后面的路径分离，返回后面的路径；

`func isASCII(s string) bool`
判断字符串 s 是否全由ASCII码构成;是，返回true；否，返回false；