package token

import (
	"context"
	"fmt"
	"testing"
	"vinesai/internel/lib/connector/constant"
	"vinesai/internel/lib/connector/env"
	"vinesai/internel/lib/connector/env/extension"
	"vinesai/internel/lib/connector/logger"
	"vinesai/internel/lib/connector/sign"
)

func TestMain(m *testing.M) {
	fmt.Println("init....")
	env.Config = env.NewEnv()
	env.Config.Init()
	extension.SetToken(constant.TUYA_TOKEN, newTokenInstance)
	extension.SetSign(constant.TUYA_SIGN, sign.NewSignWrapper)
	if logger.Log == nil {
		logger.Log = logger.NewDefaultLogger(env.Config.GetAppName(), env.Config.DebugMode())
	}
	fmt.Println("### iot core init success ###")
	m.Run()
}

func TestToken(t *testing.T) {
	tk, err := extension.GetToken(constant.TUYA_TOKEN).Do(context.Background())
	t.Log(tk, err)
}
