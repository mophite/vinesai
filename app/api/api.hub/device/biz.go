package device

import (
	"encoding/base64"

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
