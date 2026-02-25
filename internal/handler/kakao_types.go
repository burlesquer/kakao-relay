package handler

import "encoding/json"

// Kakao Webhook Request Types

type KakaoWebhookRequest struct {
	UserRequest KakaoUserRequest `json:"userRequest"`
	Bot         *KakaoBot        `json:"bot,omitempty"`
	Intent      *KakaoIntent     `json:"intent,omitempty"`
	Action      *KakaoAction     `json:"action,omitempty"`
	Contexts    []any            `json:"contexts,omitempty"`
}

type KakaoUserRequest struct {
	User        KakaoUser              `json:"user"`
	Utterance   string                 `json:"utterance"`
	CallbackURL string                 `json:"callbackUrl,omitempty"`
	Params      map[string]string      `json:"params,omitempty"`
	Block       *KakaoBlock            `json:"block,omitempty"`
}

type KakaoUser struct {
	ID         string         `json:"id"`
	Type       string         `json:"type,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

type KakaoBot struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type KakaoIntent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type KakaoAction struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Params       map[string]string `json:"params,omitempty"`
	DetailParams map[string]any    `json:"detailParams,omitempty"`
	ClientExtra  map[string]any    `json:"clientExtra,omitempty"`
}

type KakaoBlock struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Kakao Response Types

type KakaoResponse struct {
	Version     string             `json:"version"`
	Template    *KakaoTemplate     `json:"template,omitempty"`
	UseCallback bool               `json:"useCallback,omitempty"`
	Context     *KakaoContext      `json:"context,omitempty"`
	Data        map[string]any     `json:"data,omitempty"`
}

type KakaoTemplate struct {
	Outputs      []KakaoOutput      `json:"outputs"`
	QuickReplies []KakaoQuickReply  `json:"quickReplies,omitempty"`
}

type KakaoOutput struct {
	SimpleText  *KakaoSimpleText  `json:"simpleText,omitempty"`
	SimpleImage *KakaoSimpleImage `json:"simpleImage,omitempty"`
}

type KakaoSimpleText struct {
	Text string `json:"text"`
}

type KakaoSimpleImage struct {
	ImageURL string `json:"imageUrl"`
	AltText  string `json:"altText,omitempty"`
}

type KakaoQuickReply struct {
	Label       string `json:"label"`
	Action      string `json:"action"`
	MessageText string `json:"messageText,omitempty"`
}

type KakaoContext struct {
	Values []KakaoContextValue `json:"values"`
}

type KakaoContextValue struct {
	Name     string            `json:"name"`
	LifeSpan int               `json:"lifeSpan"`
	Params   map[string]string `json:"params,omitempty"`
}

// Helper functions

func NewTextResponse(text string) *KakaoResponse {
	return &KakaoResponse{
		Version: "2.0",
		Template: &KakaoTemplate{
			Outputs: []KakaoOutput{
				{SimpleText: &KakaoSimpleText{Text: text}},
			},
		},
	}
}

func NewCallbackResponse() *KakaoResponse {
	return &KakaoResponse{
		Version:     "2.0",
		UseCallback: true,
	}
}

func (r *KakaoWebhookRequest) GetPlusfriendUserKey() string {
	if r.UserRequest.User.Properties != nil {
		if key, ok := r.UserRequest.User.Properties["plusfriendUserKey"].(string); ok {
			return key
		}
	}
	return r.UserRequest.User.ID
}

func (r *KakaoWebhookRequest) GetChannelID() string {
	if r.Bot != nil && r.Bot.ID != "" {
		return r.Bot.ID
	}
	return "default"
}

func (r *KakaoWebhookRequest) ToJSON() json.RawMessage {
	data, _ := json.Marshal(r)
	return data
}
