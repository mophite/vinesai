package tuya

import (
	"testing"
	"vinesai/internel/ava"
)

func TestAgentRun(t *testing.T) {
	result, err := agentRun(ava.Background(), "打开所有灯光")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}
