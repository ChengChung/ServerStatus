package reflections

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func FullCallerFuncName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip + 1)
	return runtime.FuncForPC(pc).Name()
}

func ShortCalleeFuncName(skip int) string {
	funcWithPath := FullCallerFuncName(skip + 1)
	return strings.TrimPrefix(filepath.Ext(funcWithPath), ".")
}

func CallerFuncName(skip int) string {
	funcWithPath := FullCallerFuncName(skip + 1)
	_, funct := filepath.Split(funcWithPath)
	return funct
}

func GetTypeName(i interface{}) string {
	return fmt.Sprintf("%T", i)
}
