package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"start-feishubot/services/dify"
	"start-feishubot/services/openai"
	"strings"
	"time"

	"github.com/k0kubun/pp/v3"
)

type MessageAction struct { /*消息*/
	chatgpt *dify.Dify
}

func (m *MessageAction) Execute(a *ActionInfo) bool {
	cardId, err2 := sendOnProcess(a)
	if err2 != nil {
		return false
	}

	answer := ""
	chatResponseStream := make(chan string)
	defer close(chatResponseStream)

	done := make(chan struct{}, 3) // 添加 done 信号，保证主流程正确退出
	defer close(done)

	noContentTimeout := time.AfterFunc(20*time.Second, func() {
		pp.Println("no content timeout")
		updateFinalCard(*a.ctx, "请求超时", cardId)
		done <- struct{}{} // 发送 done 信号
	})
	defer noContentTimeout.Stop()

	msg := a.handler.sessionCache.GetMsg(*a.info.sessionId)
	msg = append(msg, openai.Messages{
		Role: "user", Content: a.info.qParsed,
	})
	conversation_id := a.handler.sessionCache.GetConversationId(*a.info.sessionId)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("panic recover", err)
				err := updateFinalCard(*a.ctx, "聊天失败", cardId)
				if err != nil {
					printErrorMessage(a, msg, err)
				}
			}
		}()

		//log.Printf("UserId: %s , Request: %s", a.info.userId, msg)

		// 这一步可能会引发panic，原因是chatResponseStream被主流程关闭，再次写入会引发panic
		if err := m.chatgpt.StreamChat(*a.ctx, a.info.qParsed, conversation_id, chatResponseStream); err != nil {
			err := updateFinalCard(*a.ctx, "聊天失败", cardId)
			if err != nil {
				printErrorMessage(a, msg, err)
			}
		}
		// 此步骤可能会引发panic
		done <- struct{}{} // 发送 done 信号
	}()

	// 开启计时器，每秒更新一次卡片
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop() // 注意在函数结束时停止 ticker

	for {
		select {
		case res, ok := <-chatResponseStream:
			if !ok {
				return false
			}
			noContentTimeout.Stop()
			answer += res
			//pp.Println("answer", answer)
		case <-ticker.C: //
			err := updateTextCard(*a.ctx, answer, cardId)
			if err != nil {
				printErrorMessage(a, msg, err)
				return false
			}
		case <-done: // 添加 done 信号的处理
			err := updateFinalCard(*a.ctx, answer, cardId)
			if err != nil {
				printErrorMessage(a, msg, err)
				return false
			}
			msg := append(msg, openai.Messages{
				Role: "assistant", Content: answer,
			})
			a.handler.sessionCache.SetMsg(*a.info.sessionId, msg)

			log.Printf("\n\n\n")
			log.Printf("Success request: UserId: %s , Request: %s , Response: %s", a.info.userId, msg, answer)
			jsonByteArray, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling JSON request: UserId: %s , Request: %s , Response: %s", a.info.userId, jsonByteArray, answer)
			}
			jsonStr := strings.ReplaceAll(string(jsonByteArray), "\\n", "")
			jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
			log.Printf("\n\n\n")
			log.Printf("Success request plain jsonStr: UserId: %s , Request: %s , Response: %s",
				a.info.userId, jsonStr, answer)
			return false
		}
	}
}

func printErrorMessage(a *ActionInfo, msg []openai.Messages, err error) {
	log.Printf("Failed request: UserId: %s , Request: %s , Err: %s", a.info.userId, msg, err)
}

func sendOnProcess(a *ActionInfo) (*string, error) {
	// send 正在处理中
	cardId, err := sendOnProcessCard(*a.ctx, a.info.sessionId, a.info.msgId)
	if err != nil {
		return nil, err
	}
	return cardId, nil

}
