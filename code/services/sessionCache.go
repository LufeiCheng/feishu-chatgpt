package services

import (
	"start-feishubot/services/openai"
	"time"

	"github.com/patrickmn/go-cache"
)

type SessionMode string
type SessionService struct {
	cache *cache.Cache
}
type PicSetting struct {
	resolution Resolution
}
type Resolution string

type SessionMeta struct {
	Mode           SessionMode       `json:"mode"`
	Msg            []openai.Messages `json:"msg,omitempty"`
	PicSetting     PicSetting        `json:"pic_setting,omitempty"`
	AIMode         openai.AIMode     `json:"ai_mode,omitempty"`
	ConversationId string            `json:"conversation_id,omitempty"`
}

const (
	Resolution256  Resolution = "256x256"
	Resolution512  Resolution = "512x512"
	Resolution1024 Resolution = "1024x1024"
)
const (
	ModePicCreate SessionMode = "pic_create"
	ModePicVary   SessionMode = "pic_vary"
	ModeGPT       SessionMode = "gpt"
)

type SessionServiceCacheInterface interface {
	Get(sessionId string) *SessionMeta
	Set(sessionId string, sessionMeta *SessionMeta)
	GetMsg(sessionId string) []openai.Messages
	SetMsg(sessionId string, msg []openai.Messages)
	SetMode(sessionId string, mode SessionMode)
	GetMode(sessionId string) SessionMode
	GetAIMode(sessionId string) openai.AIMode
	SetAIMode(sessionId string, aiMode openai.AIMode)
	SetConversationId(sessionId string, conversationId string)
	GetConversationId(sessionId string) string
	SetPicResolution(sessionId string, resolution Resolution)
	GetPicResolution(sessionId string) string
	Clear(sessionId string)
}

var sessionServices *SessionService

// implement Get interface
func (s *SessionService) Get(sessionId string) *SessionMeta {
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return nil
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return sessionMeta
}

// implement Set interface
func (s *SessionService) Set(sessionId string, sessionMeta *SessionMeta) {
	maxCacheTime := time.Hour * 12
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

func (s *SessionService) GetMode(sessionId string) SessionMode {
	// Get the session mode from the cache.
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return ModeGPT
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return sessionMeta.Mode
}

func (s *SessionService) SetMode(sessionId string, mode SessionMode) {
	maxCacheTime := time.Hour * 12
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		sessionMeta := &SessionMeta{Mode: mode}
		s.cache.Set(sessionId, sessionMeta, maxCacheTime)
		return
	}
	sessionMeta := sessionContext.(*SessionMeta)
	sessionMeta.Mode = mode
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

func (s *SessionService) GetAIMode(sessionId string) openai.AIMode {
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return openai.Balance
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return sessionMeta.AIMode
}

// Get conversation id from the cache.
func (s *SessionService) GetConversationId(sessionId string) string {
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return ""
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return sessionMeta.ConversationId
}

// Set conversation id to the cache.
func (s *SessionService) SetConversationId(sessionId string, conversationId string) {
	maxCacheTime := time.Hour * 12
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		sessionMeta := &SessionMeta{ConversationId: conversationId}
		s.cache.Set(sessionId, sessionMeta, maxCacheTime)
		return
	}
	sessionMeta := sessionContext.(*SessionMeta)
	sessionMeta.ConversationId = conversationId
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

// SetAIMode set the ai mode for the session.
func (s *SessionService) SetAIMode(sessionId string, aiMode openai.AIMode) {
	maxCacheTime := time.Hour * 12
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		sessionMeta := &SessionMeta{AIMode: aiMode}
		s.cache.Set(sessionId, sessionMeta, maxCacheTime)
		return
	}
	sessionMeta := sessionContext.(*SessionMeta)
	sessionMeta.AIMode = aiMode
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

func (s *SessionService) GetMsg(sessionId string) (msg []openai.Messages) {
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return nil
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return sessionMeta.Msg
}

func (s *SessionService) SetMsg(sessionId string, msg []openai.Messages) {
	maxLength := 4096
	maxCacheTime := time.Hour * 12

	//限制对话上下文长度
	for getStrPoolTotalLength(msg) > maxLength {
		msg = append(msg[:1], msg[2:]...)
	}

	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		sessionMeta := &SessionMeta{Msg: msg}
		s.cache.Set(sessionId, sessionMeta, maxCacheTime)
		return
	}
	sessionMeta := sessionContext.(*SessionMeta)
	sessionMeta.Msg = msg
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

func (s *SessionService) SetPicResolution(sessionId string,
	resolution Resolution) {
	maxCacheTime := time.Hour * 12

	//if not in [Resolution256, Resolution512, Resolution1024] then set
	//to Resolution256
	switch resolution {
	case Resolution256, Resolution512, Resolution1024:
	default:
		resolution = Resolution256
	}

	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		sessionMeta := &SessionMeta{PicSetting: PicSetting{resolution: resolution}}
		s.cache.Set(sessionId, sessionMeta, maxCacheTime)
		return
	}
	sessionMeta := sessionContext.(*SessionMeta)
	sessionMeta.PicSetting.resolution = resolution
	s.cache.Set(sessionId, sessionMeta, maxCacheTime)
}

func (s *SessionService) GetPicResolution(sessionId string) string {
	sessionContext, ok := s.cache.Get(sessionId)
	if !ok {
		return string(Resolution256)
	}
	sessionMeta := sessionContext.(*SessionMeta)
	return string(sessionMeta.PicSetting.resolution)

}

func (s *SessionService) Clear(sessionId string) {
	// Delete the session context from the cache.
	s.cache.Delete(sessionId)
}

func GetSessionCache() SessionServiceCacheInterface {
	if sessionServices == nil {
		sessionServices = &SessionService{cache: cache.New(time.Hour*12, time.Hour*1)}
	}
	return sessionServices
}

func getStrPoolTotalLength(strPool []openai.Messages) int {
	var total int
	for _, v := range strPool {
		total += v.CalculateTokenLength()
	}
	return total
}
