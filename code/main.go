package main

import (
	"context"
	"encoding/json"
	"log"

	"start-feishubot/handlers"
	"start-feishubot/initialization"
	"start-feishubot/services/openai"

	"github.com/gin-gonic/gin"
	sdkginext "github.com/larksuite/oapi-sdk-gin"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/pflag"
)

func main() {
	initialization.InitRoleList()
	pflag.Parse()

	config := initialization.GetConfig()
	// 打印一下实际读取到的配置
	globalConfigPrettyString, _ := json.MarshalIndent(config, "", "    ")
	log.Println(string(globalConfigPrettyString))

	initialization.LoadLarkClient(*config)
	initialization.InitLogger(*config)
	gpt := openai.NewChatGPT(*config)
	handlers.InitHandlers(gpt, *config)

	eventHandler := dispatcher.NewEventDispatcher(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey).
		OnP2MessageReceiveV1(handlers.Handler).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			return handlers.ReadHandler(ctx, event)
		})

	cardHandler := larkcard.NewCardActionHandler(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey,
		handlers.CardHandler())

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/webhook/event",
		sdkginext.NewEventHandlerFunc(eventHandler))
	r.POST("/webhook/card",
		sdkginext.NewCardActionHandlerFunc(
			cardHandler))

	if err := initialization.StartServer(*config, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
