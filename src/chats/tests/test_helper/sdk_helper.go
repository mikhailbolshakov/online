package test_helper

import (
	"chats/sdk"
	"time"
)

const (
	url     = "nats://localhost:4222"
	timeout = time.Duration(1000) * time.Millisecond
	subject = "go-sdk-test"
	token   = ""
	topic	= "online.1.0"
	ws		= "ws://localhost:8000/ws/?token="
)

func InitSdk() (*sdk.Sdk, error) {
	opt := &sdk.Options{url, token, timeout, true}
	return sdk.Init(opt)
}