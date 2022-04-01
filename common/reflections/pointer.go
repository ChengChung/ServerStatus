package reflections

import (
	"fmt"
	"reflect"
)

func DereferenceIfPtr(arg interface{}) (reflect.Value, bool) {
	if reflect.TypeOf(arg).Kind() != reflect.Ptr {
		return reflect.ValueOf(nil), false
	}
	ptr := reflect.Indirect(reflect.ValueOf(arg))
	return ptr, true
}

//	检查pointer是否为prototype类型的指针
//	入参为空，pointer不是指针，指针为空指针都会返回错误
func CheckPointerInnerType(prototype interface{}, pointer interface{}) (bool, error) {
	if prototype == nil || pointer == nil {
		return false, fmt.Errorf("%s expected non-nil interface content", CallerFuncName(0))
	}
	ptrObj, ok := DereferenceIfPtr(pointer)
	if !ok {
		return false, fmt.Errorf("%s expected pointer type of argument %s", CallerFuncName(0), GetTypeName(pointer))
	}
	if ptrObj == reflect.ValueOf(nil) {
		return false, fmt.Errorf("%s expected non-nil pointer", CallerFuncName(0))
	}
	if reflect.TypeOf(prototype) != ptrObj.Type() {
		return false, nil
	}

	return true, nil
}

func SetPointerValue(val interface{}, target interface{}) {
	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(val))
}
