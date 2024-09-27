package sign

import (
	"context"
	"testing"
	"vinesai/internel/lib/connector/constant"
	"vinesai/internel/lib/connector/env"
)

func TestSign(t *testing.T) {
	env.Config = env.NewEnv()
	env.Config.Init()
	ctx := context.Background()
	ctx = context.WithValue(ctx, constant.TOKEN, "123")
	ctx = context.WithValue(ctx, constant.TS, "123")
	ctx = context.WithValue(ctx, constant.NONCE, "123")
	sw := &signWrapper{}
	t.Log(sw.Sign(ctx))
}
