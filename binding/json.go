package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

type jsonBinding struct {
	IsValidate            bool
	DisallowUnknownFields bool
}

func (jsonBinding) Name() string {
	return "json"
}

// Bind JSON绑定器
func (b jsonBinding) Bind(r *http.Request, obj any) error {
	body := r.Body // 传参的内容放在 http.Request.Body 中
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body) // ** 解码器中含有body中的内容 **
	if b.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if b.IsValidate {
		if err := validateParam(obj, decoder); err != nil {
			return err
		}
	} else {
		err := decoder.Decode(obj)
		if err != nil {
			return err
		}
	}
	return validate(obj)
}

// ** 好好学习反射 **
func validateParam(obj any, decoder *json.Decoder) error {
	// 反射
	valueOf := reflect.ValueOf(obj)
	// 判断其是否为指针类型
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("This argument must have a pointer type ")
	}
	temp := valueOf.Elem().Interface()
	vOf := reflect.ValueOf(temp)
	// obj 类型判断后才能进行解析操作
	switch vOf.Kind() {
	case reflect.Struct:
		return checkParamDecoder(vOf, obj, decoder)
	case reflect.Slice, reflect.Array:
		elem := vOf.Type().Elem()
		if elem.Kind() == reflect.Struct {
			return checkParamSliceDecoder(elem, obj, decoder)
		} else if elem.Kind() == reflect.Pointer {
			return checkParamSliceDecoder(elem.Elem(), obj, decoder)
		}
	default:
		if err := decoder.Decode(obj); err != nil {
			return err
		}
	}
	return nil
}

func checkParamDecoder(vOf reflect.Value, obj any, decoder *json.Decoder) error {
	// 解析为map，然后根据map中的key进行对比
	mapData := make(map[string]any)
	// 因为解码器中带有body内容，所以可以将body解析成对应map
	if err := decoder.Decode(&mapData); err != nil {
		log.Println(err)
		return err
	}
	for i := 0; i < vOf.NumField(); i++ {
		field := vOf.Type().Field(i)
		name := field.Name                // post form 中的 json 属性名
		jsonName := field.Tag.Get("json") // 寻找结构体中`json:`的属性名
		required := field.Tag.Get("spxgo")
		val := mapData[jsonName]
		if val == nil && required == "required" {
			return errors.New(
				fmt.Sprintf("filed [%s] is requird", name))
		}
	}
	// 因为decoder.Decode 不能重复执行 下面的操作相当于decoder.Decode(obj)
	b, err := json.Marshal(mapData) // 将数据编码成json字符串
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, obj); err != nil { // 将json字符串解码到相应的数据结构 obj相当于是一个数据结构模型
		return err
	}
	return err
}

func checkParamSliceDecoder(elem reflect.Type, obj any, decoder *json.Decoder) error {
	mapData := make([]map[string]any, 0)
	if err := decoder.Decode(&mapData); err != nil {
		return err
	}
	if len(mapData) <= 0 {
		return nil
	}
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		jsonName := field.Tag.Get("json")
		required := field.Tag.Get("spxgo")
		for _, v := range mapData {
			if v[jsonName] == nil && required == "required" {
				return errors.New(
					fmt.Sprintf("field [%s] is required", jsonName))
			}
		}
	}
	b, err := json.Marshal(mapData) // 将数据编码成json字符串
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, obj); err != nil { // 将json字符串解码到相应的数据结构
		return err
	}
	return err
}
