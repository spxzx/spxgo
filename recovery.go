package spxgo

import (
	"errors"
	"fmt"
	"gitbuh.com/spxzx/spxgo/serror"
	"net/http"
	"runtime"
	"strings"
)

func detailMsg(err any) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v", err))
	for _, pc := range pcs[0:n] {
		forPC := runtime.FuncForPC(pc)
		file, line := forPC.FileLine(pc)
		sb.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return sb.String()
}

func Recovery(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				err2 := err.(error)
				if err2 != nil {
					var spxError *serror.SpxError
					if errors.As(err2, &spxError) {
						spxError.ExecuteResult()
						return
					}
				}
				// 在debug bindJson 的时候发生了意料之外的情况，捕获了错误，但是并没有走下面的代码
				c.Logger.Error(detailMsg(err))
				c.Fail(http.StatusInternalServerError, "500 Internal Server Error")
			}
		}()
		next(c)
	}
}
