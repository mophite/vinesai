package lib

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"vinesai/internel/ava"
)

var (
	Client *http.Client
)

func init() {
	Client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			//TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
			//Proxy:             http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,  // tcp连接超时时间
				KeepAlive: 600 * time.Second, // 保持长连接的时间
			}).DialContext, // 设置连接的参数
			MaxIdleConns:          50,                // 最大空闲连接
			MaxConnsPerHost:       100,               //每个host建立多少个连接
			MaxIdleConnsPerHost:   100,               // 每个host保持的空闲连接数
			ExpectContinueTimeout: 60 * time.Second,  // 等待服务第一响应的超时时间
			IdleConnTimeout:       600 * time.Second, // 空闲连接的超时时间
		},
	}
}

// CheckRespStatus 状态检查
func CheckRespStatus(resp *http.Response) ([]byte, error) {
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return bodyBytes, nil
	}
	return nil, errors.New(string(bodyBytes))
}

func POST(c *ava.Context, url string, data []byte, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := Client.Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}

func Get(c *ava.Context, url string, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := Client.Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}

func GetWithout(url string, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		ava.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := Client.Do(request)
	if err != nil {
		ava.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		ava.Error(err)
		return nil, err
	}

	return rsp, nil
}

func PUT(c *ava.Context, url string, data []byte, header map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		c.Error(err)
		return nil, err
	}

	if header != nil {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	resp, err := Client.Do(request)
	if err != nil {
		c.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	rsp, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Error(err)
		return nil, err
	}

	return rsp, nil
}
