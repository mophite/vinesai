package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtKey = []byte("DOGQ6MNVIU9Y5J5LK0PWB1A8H2Z4ERCB")

type tokenClaims struct {
	Timestamp   string
	RedirectUrl string
	jwtv5.RegisteredClaims
}

func parseJWToken(token string) (*tokenClaims, error) {
	t, err := jwtv5.ParseWithClaims(token, &tokenClaims{}, func(t *jwtv5.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*tokenClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

var defaultExpiryDelta = time.Hour * 24 * 30 * 12 * 100

func generateJWToken(c *ava.Context, phone, redirect string) (string, int64) {
	expiry := jwtv5.NewNumericDate(x.LocalTimeNow().Add(defaultExpiryDelta))
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, tokenClaims{
		Timestamp:   x.LocalTimeNow().Format(time.RFC3339),
		RedirectUrl: redirect,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    "vinesai",
			Subject:   "oauth2.0授权",
			Audience:  []string{phone},
			ExpiresAt: expiry,
			NotBefore: jwtv5.NewNumericDate(x.LocalTimeNow()), //token在此时间之前不能被接收处理
			IssuedAt:  jwtv5.NewNumericDate(x.LocalTimeNow()),
			ID:        uuid.New().String(),
		},
	})

	str, err := token.SignedString(jwtKey)
	if err != nil {
		c.Errorf("generateJWToken |err=%v", err)
		return "", 0
	}

	return str, expiry.Unix()
}

// 创建docker容器
func createDocker(c *ava.Context, port string) error {
	ctx := context.Background()

	// 创建绑定源路径
	sourcePath := fmt.Sprintf("/data/homeassistant/%s/config", port)
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create source path: %w", err)
	}

	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.Error(err)
		return err
	}

	// 定义端口映射
	exposedPorts, portBindings, err := nat.ParsePortSpecs([]string{fmt.Sprintf("%s:8123", port)})
	if err != nil {
		c.Error(err)
		return err
	}

	// 创建容器配置
	containerConfig := &container.Config{
		Image:        "homeassistant/home-assistant:latest",
		Env:          []string{"TZ=Asia/Shanghai"},
		ExposedPorts: exposedPorts,
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		//StorageOpt: map[string]string{
		//	"size": "1G",
		//},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: sourcePath,
				Target: "/config",
			},
		},
		PortBindings: portBindings,
	}

	// 创建网络配置
	networkConfig := &network.NetworkingConfig{}

	// 创建并启动容器
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, "homeassistant_"+port)
	if err != nil {
		c.Error(err)
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		c.Error(err)
		return err
	}

	fmt.Printf("Container %s created and started\n", resp.ID)

	return nil
}
