package models

type WechatNotificationResponse struct {
	ErrorMessage string `json:"errmsg"`
	ErrorCode    int    `json:"errcode"`
	AccessToken  string    `json:"access_token"`
}

type WechatNotification struct {
	ChatId      string                  `json:"chatid"`
	MessageType string                  `json:"msgtype"`
	Markdown    *WechatNotificationText `json:"markdown,omitempty"`
	Safe        int8                    `json:"safe"`
}

type WechatNotificationText struct {
	Content string `json:"content"`
}

type WechatCorpInfo struct {
	CorpChatId string
	CorpId     string
	CorpSecret string
	WechatUrl  string
}
