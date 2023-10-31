package device

import (
	"encoding/base64"
	"fmt"
	"strings"

	v20190614 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

var gClient *v20190614.Client

func init() {
	var err error

	credential := common.NewCredential(
		"AKIDwlV9eZN3b0Hd0HTkFoLqBWTEnEZmlZLl",
		"IL8iIsKZVL0a6TxM7j7SggRFvcIqJiyB",
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "asr.tencentcloudapi.com"
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10

	gClient, err = v20190614.NewClient(credential, regions.Beijing, cpf)

	if err != nil {
		panic(err)
	}
}

func asr(binary []byte, engSerViceType string) (*v20190614.SentenceRecognitionResponse, error) {

	if engSerViceType == "" {
		engSerViceType = "8k_zh"
	}
	var request = v20190614.NewSentenceRecognitionRequest()
	request.EngSerViceType = common.StringPtr(engSerViceType)
	request.SourceType = common.Uint64Ptr(1)
	request.VoiceFormat = common.StringPtr("wav")
	request.Data = common.StringPtr(base64.StdEncoding.EncodeToString(binary))
	request.DataLen = common.Int64Ptr(int64(len(binary)))
	request.SubServiceType = common.Uint64Ptr(2)

	return gClient.SentenceRecognition(request)
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
