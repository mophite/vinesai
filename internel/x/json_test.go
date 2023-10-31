package x

import (
	"fmt"
	"strings"
	"testing"
)

var s = `{"message":"Text processed successfully","result":"AiavaControl:###{‘1004’:{‘setSwitch’:‘true’},‘1002’:{‘setCoolingSetpoint’:‘20’,‘setThermostatMode’:‘off’},‘1003’:{‘setSwitch’:‘false’},‘1005’:{‘setSwitch’:‘true’}}&&&&& 【将灯光开启，将空调关闭，将空气净化器关闭，将电热水器开启】 <<<好的，主人，我会立即执行您的命令。>>>","code":200}`
var ss = `{"name":"test"}`

func TestMustMarshal(t *testing.T) {
	s = strings.ReplaceAll(s, "‘", "'")
	s = strings.ReplaceAll(s, "’", "'")
	//s = strings.ReplaceAll(s, `"`, `\"`)
	//fmt.Println(s)
	var tt = struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Result  string `json:"result"`
	}{}
	err := MustNativeUnmarshal([]byte(s), &tt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("---", tt.Result)
	e, i, err := parseRobotCom(tt.Result)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(e)
	fmt.Println(i)
}

const magicStr = "AiavaControl:###"
const magicStr2 = "AiavaControl：###"

func parseRobotCom(s string) (string, string, error) {

	var mgic = magicStr
	i := strings.Index(s, magicStr)
	if i < 0 {
		mgic = magicStr2
		i := strings.Index(s, magicStr2)
		if i < 0 {
			return "", "", fmt.Errorf("err")
		}
	}

	j := strings.Index(s, "&&&&&")
	if j < 1 {
		return "", "", fmt.Errorf("%s err", s)
	}

	h := strings.Index(s, "<<<")
	he := strings.Index(s, ">>>")
	if h < 0 || he < 0 {
		return "", "", fmt.Errorf("%s <> err", s)
	}

	d := s[h+3 : he]

	s = s[:h]

	s = strings.Replace(s, mgic, "", -1)
	s = strings.Replace(s, "&&&&&", "", -1)
	s = strings.Replace(s, "【", "", 1)
	s = strings.Replace(s, "】", "", 1)
	return s, d, nil
}
