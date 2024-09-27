package message

import (
	"fmt"
	"testing"
	"time"
	"vinesai/internel/lib/connector/constant"
	"vinesai/internel/lib/connector/env"
	"vinesai/internel/lib/connector/env/extension"
	"vinesai/internel/lib/connector/logger"
	"vinesai/internel/lib/connector/message/event"
)

func TestMain(m *testing.M) {
	fmt.Println("init....")
	env.Config = env.NewEnv()
	env.Config.Init()
	extension.SetMessage(constant.TUYA_MESSAGE, newMessageInstance)
	if logger.Log == nil {
		logger.Log = logger.NewDefaultLogger(env.Config.GetAppName(), env.Config.DebugMode())
	}
	fmt.Println("### iot core init success ###")
	m.Run()
}

func TestEventMsg(t *testing.T) {
	extension.GetMessage(constant.TUYA_MESSAGE).InitMessageClient()
	extension.GetMessage(constant.TUYA_MESSAGE).SubEventMessage(func(m *event.NameUpdateMessage) {
		logger.Log.Info("=========== name update： ==========")
		logger.Log.Info(m)
	})
	extension.GetMessage(constant.TUYA_MESSAGE).SubEventMessage(func(m *event.StatusReportMessage) {
		logger.Log.Info("=========== report data： ==========")
		for _, v := range m.Status {
			logger.Log.Info(v.Code, v.Value)
		}
	})

	time.Sleep(20 * time.Second)
	extension.GetMessage(constant.TUYA_MESSAGE).Stop()
	t.Log("end.....")
}
