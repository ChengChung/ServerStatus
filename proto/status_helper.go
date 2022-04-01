package proto

import (
	"github.com/chengchung/ServerStatus/common/reflections"
)

func init() {
	reflections.RegisterTypeMapping(ServerStatus{})
}

func (status *ServerStatus) SetTypeValue(jsonField string, value interface{}) {
	reflections.SetTypeValue(status, jsonField, value)
}
