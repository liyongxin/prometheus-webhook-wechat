package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/liyongxin/prometheus-webhook-wechat/models"
	"github.com/liyongxin/prometheus-webhook-wechat/template"
	"github.com/liyongxin/prometheus-webhook-wechat/webrouter"
	"fmt"
)

func BuildWechatNotification(rs *webrouter.WechatResource, promMessage *models.WebhookMessage) (*models.WechatNotification, error) {
	title, err := template.ExecuteTextString(`{{ template "wechat.link.title" . }}`, promMessage)
	if err != nil {
		return nil, err
	}
	content, err := template.ExecuteTextString(`{{ template "wechat.link.content" . }}`, promMessage)
	if err != nil {
		return nil, err
	}

	var notifyContent = models.WechatNotificationText{
		Content: title + content,
	}

	notification := &models.WechatNotification{
		ChatId: rs.CorpChatId,
		MessageType: "markdown",
		Markdown: &notifyContent,
	}
	return notification, nil
}

func SendWechatNotification(httpClient *http.Client, rs *webrouter.WechatResource, notification *models.WechatNotification) (*models.WechatNotificationResponse, error) {
	body, err := json.Marshal(&notification)
	if err != nil {
		return nil, errors.Wrap(err, "error encoding Wechat request")
	}
	tokenResp, err := getAccessToken(rs)
	if err != nil {
		return nil, errors.Wrap(err, "error get access_token")
	}
	url := fmt.Sprintf("%s/appchat/send?access_token=%s", rs.WechatUrl, tokenResp.AccessToken)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "error building Wechat request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	req, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "error sending notification to Wechat")
	}
	defer req.Body.Close()

	if req.StatusCode != 200 {
		return nil, errors.Errorf("unacceptable response code %d", req.StatusCode)
	}

	var robotResp models.WechatNotificationResponse
	enc := json.NewDecoder(req.Body)
	if err := enc.Decode(&robotResp); err != nil {
		return nil, errors.Wrap(err, "error decoding response from Wechat")
	}

	return &robotResp, nil
}

func getAccessToken(rs *webrouter.WechatResource) (*models.WechatNotificationResponse, error) {
	url := fmt.Sprintf("%s/gettoken?corpid=%s&corpsecret=%s", rs.WechatUrl, rs.CorpId, rs.CorpSecret)
	httpRes, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "error building Wechat request")
	}
	defer httpRes.Body.Close()
	if httpRes.StatusCode != 200 {
		return nil, errors.Errorf("unacceptable response code %d", httpRes.StatusCode)
	}
	var robotResp models.WechatNotificationResponse
	enc := json.NewDecoder(httpRes.Body)
	if err := enc.Decode(&robotResp); err != nil {
		return nil, errors.Wrap(err, "error decoding response from Wechat")
	}

	return &robotResp, nil
}