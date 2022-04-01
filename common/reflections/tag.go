package reflections

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

var type_mappings map[string]map[string]int = make(map[string]map[string]int)

func RegisterTypeMapping(value interface{}) {
	type_mappings[reflect.PtrTo(reflect.TypeOf(value)).String()] = GetTypeJsonMapping(value)
}

func GetTypeJsonMapping(value interface{}) map[string]int {
	res := make(map[string]int)

	reflectType := reflect.TypeOf(value)
	num := reflectType.NumField()

	//	not recursive
	for i := 0; i < num; i++ {
		f := reflectType.Field(i)
		name := f.Tag.Get("json")
		res[name] = i
	}

	return res
}

func SetTypeValue(object_ptr interface{}, jsonField string, value interface{}) {
	ntype := reflect.TypeOf(object_ptr).String()
	if m, ok := type_mappings[ntype]; !ok {
		logrus.Errorf("fail to find type %s", ntype)
	} else {
		if idx, ok := m[jsonField]; !ok {
			logrus.Error("fail to find field ", jsonField)
		} else {
			f := reflect.Indirect(reflect.ValueOf(object_ptr)).Field(idx)
			switch f.Kind() {
			case reflect.String:
				f.SetString(value.(string))
			case reflect.Float64:
				f.SetFloat(value.(float64))
			case reflect.Int64:
				f.SetInt(value.(int64))
			default:
				logrus.Error("unsupported value type ", f.Kind())
			}
		}
	}
}
