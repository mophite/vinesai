package device

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestAsr(t *testing.T) {
	f, err := os.Open("./test.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fd, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := asr(fd)
	rsp.ToJsonString()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(*rsp.Response.Result)
}
