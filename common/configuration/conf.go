package configuration

import (
	"encoding/json"
	"fmt"

	"github.com/chengchung/ServerStatus/common/reflections"
)

func ParseConf(rawcfg interface{}, target interface{}) error {
	if ok, err := reflections.CheckPointerInnerType(rawcfg, target); err != nil {
		return err
	} else if ok {
		reflections.SetPointerValue(rawcfg, target)
		return nil
	}

	if jsoncfg, ok := rawcfg.(json.RawMessage); !ok {
		return fmt.Errorf("mismatched type error, expected %s got %s", reflections.GetTypeName(target), reflections.GetTypeName(rawcfg))
	} else if err := json.Unmarshal(jsoncfg, target); err != nil {
		return err
	}

	return nil
}
