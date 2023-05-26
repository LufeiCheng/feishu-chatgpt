package dify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"start-feishubot/initialization"
	customOpenai "start-feishubot/services/openai"
)

type Messages struct {
	Inputs         map[string]string `json:"inputs"`
	Query          string            `json:"query"`
	ResponseMode   string            `json:"response_mode"`
	ConversationId string            `json:"conversation_id"`
	User           string            `json:"user"`
}

/*
	{
	    "event": "message",
	    "task_id": "b97116ed-a826-46e7-9cf2-2a0a03c10815",
	    "id": "fd2f0589-465f-4e56-840a-6c0906032dcd",
	    "answer": "",
	    "created_at": 1685084584,
	    "conversation_id": "16573d51-3037-4784-ac2a-1c3a9301623f"
	}
*/
type Response struct {
	Event          string `json:"event"`
	TaskId         string `json:"task_id"`
	Id             string `json:"id"`
	Answer         string `json:"answer"`
	CreatedAt      int    `json:"created_at"`
	ConversationId string `json:"conversation_id"`
}

type Dify struct {
	config *initialization.Config
}

func NewDify(config *initialization.Config) *Dify {
	return &Dify{config: config}
}

func (c *Dify) StreamChat(ctx context.Context,
	query string, converstationId string,
	responseStream chan string) error {

	// generate msg
	msg := Messages{
		Inputs:         map[string]string{},
		Query:          query,
		ResponseMode:   "streaming",
		ConversationId: converstationId,
		User:           "roger",
	}

	return c.StreamChatWithHistory(ctx, msg, 2000,
		responseStream)
}

func (c *Dify) StreamChatWithHistory(ctx context.Context, msg Messages, maxTokens int,
	responseStream chan string,
) error {
	url := "http://152.32.168.165:8008/v1/chat-messages"
	method := "POST"
	requestBodyData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	client, err := customOpenai.GetProxyClient(c.config.HttpProxy)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(requestBodyData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+"app-HyiTBvlsjEkes86YTIvhcICQ")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("CreateCompletionStream returned error: %v\n", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}
		var resBody *Response
		err := json.Unmarshal(data[6:], &resBody)
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return err
		}
		responseStream <- resBody.Answer
	}
	return nil
}
