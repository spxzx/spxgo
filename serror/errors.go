package serror

type SpxError struct {
	err       error
	ErrorFunc ErrorFunc
}

func Default() *SpxError {
	return &SpxError{}
}

func (e *SpxError) Error() string {
	return e.err.Error()
}

func (e *SpxError) Put(err error) {
	e.check(err)
}

func (e *SpxError) check(err error) {
	if err != nil {
		e.err = err
		panic(e)
	}
}

type ErrorFunc func(spxError *SpxError)

// Result 暴露一个方法让用户自定义
func (e *SpxError) Result(ef ErrorFunc) {
	e.ErrorFunc = ef
}

func (e *SpxError) ExecuteResult() {
	e.ErrorFunc(e)
}
