package binding

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

var Validator StructValidator = &defaultValidator{}

// 引入第三方校验

type StructValidator interface {
	ValidateStruct(any) error
	Engine() any
}

type SliceValidationError []error

func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			_, _ = fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					_, _ = fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

type defaultValidator struct {
	one      sync.Once // 单例
	validate *validator.Validate
}

// lazyInit 懒加载
func (d *defaultValidator) lazyInit() {
	d.one.Do(func() {
		d.validate = validator.New()
	})
}

func (d *defaultValidator) validateStruct(obj any) error {
	d.lazyInit()
	return d.validate.Struct(obj)
}

func (d *defaultValidator) ValidateStruct(obj any) error {
	valueOf := reflect.ValueOf(obj)
	switch valueOf.Kind() {
	case reflect.Pointer:
		return d.ValidateStruct(valueOf.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := valueOf.Len()
		sliceValidationError := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(valueOf.Index(i).Interface()); err != nil {
				sliceValidationError = append(sliceValidationError, err)
			}
		}
		if len(sliceValidationError) == 0 {
			return nil
		}
		return sliceValidationError // 因为重写了Error() 所以返回这个不报错
	}
	return nil
}

// Engine 作用暂时不明 猜测：懒加载一次validator后(这样有了具体的validator?)将其返回道有需要的地方
func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

func validate(obj any) error {
	return Validator.ValidateStruct(obj)
}
