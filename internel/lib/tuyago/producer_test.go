package tuyago

import (
	"testing"
	"vinesai/internel/ava"
)

func TestProducer(t *testing.T) {
	var p = &protocolData{
		Protocol: 45,
		T:        1678693809396,
		Data: `{
    "bizCode":"textToSpeech",
    "bizData":{
        "brandCode":"abc*******",
        "voiceId":"def*******",
        "command": [{
            "intent": "welcome",
            "content":{
                "value": "欢迎入住本酒店"
            }
        }]
    },
    "ts":1636682568127
}`,
		//Sign:     "",
	}

	Producer(ava.Background(), p)
}
