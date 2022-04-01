package configuration

import (
	"encoding/json"
	"fmt"
	"testing"
)

type Conf1 struct {
	Val   string            `json:"val"`
	Field string            `json:"field"`
	Items map[string]string `json:"items"`
}

type Conf2 struct {
	Val   int            `json:"val"`
	Field int            `json:"field"`
	Items map[string]int `json:"items"`
}

func TestParseConf(t *testing.T) {
	conf1 := Conf1{
		Val:   "abc",
		Field: "cba",
		Items: map[string]string{"k": "1"},
	}

	var conf1b Conf1
	if err := ParseConf(conf1, &conf1b); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", conf1b)

	var conf1c *Conf1
	if err := ParseConf(conf1, conf1c); err == nil {
		t.Fatal("expect error when parsing conf1c")
	} else {
		fmt.Println(err)
	}

	var conf1d Conf2
	if err := ParseConf(conf1, &conf1d); err == nil {
		t.Fatal("expect error when parsing conf1c")
	} else {
		fmt.Println(err)
	}

	var conf1e Conf1
	if err := ParseConf(conf1, conf1e); err == nil {
		t.Fatal("expect error when parsing conf1c")
	} else {
		fmt.Println(err)
	}

	var conf1h Conf1
	json1 := json.RawMessage(`{"val":"555", "field":"666", "items":{"a":"b"}}`)
	if err := ParseConf(json1, &conf1h); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", conf1h)

	var conf1i Conf1
	json2 := json.RawMessage(`{"val":555, "field":666, "items":{"a":5}}`)
	if err := ParseConf(json2, &conf1i); err == nil {
		t.Fatal("expect error when parsing conf1c")
	} else {
		fmt.Println(err)
	}

	var conf2a Conf2
	if err := ParseConf(json2, &conf2a); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", conf2a)

	if err := ParseConf(nil, nil); err == nil {
		t.Fatal("expect error when parsing conf1c")
	} else {
		fmt.Println(err)
	}
}
