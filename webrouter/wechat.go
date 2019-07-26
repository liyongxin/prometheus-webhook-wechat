package webrouter

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/liyongxin/prometheus-webhook-wechat/models"
	"github.com/liyongxin/prometheus-webhook-wechat/notifier"
)

type WechatResource struct {
	Logger     log.Logger
	CorpChatId string
	CorpId     string
	CorpSecret string
	WechatUrl  string
	HttpClient *http.Client
}

func (rs *WechatResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/general_alerting/send", rs.SendNotification)
	return r
}

func (rs *WechatResource) SendNotification(w http.ResponseWriter, r *http.Request) {
	logger := rs.Logger
	corpChatId := rs.CorpChatId
	corpId := rs.CorpId
	corpSecret := rs.CorpSecret
	if corpChatId == "" || corpId == "" || corpSecret == "" {
		http.Error(w, "Missing required params", http.StatusBadRequest)
		http.NotFound(w, r)
		return
	}

	var promMessage models.WebhookMessage
	if err := json.NewDecoder(r.Body).Decode(&promMessage); err != nil {
		level.Error(logger).Log("msg", "Cannot decode prometheus webhook JSON request", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	notification, err := notifier.BuildWechatNotification(rs, &promMessage)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to build notification", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	robotResp, err := notifier.SendWechatNotification(rs.HttpClient, rs, notification)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to send notification", "err", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if robotResp.ErrorCode != 0 {
		level.Error(logger).Log("msg", "Failed to send notification to Wechat", "respCode", robotResp.ErrorCode, "respMsg", robotResp.ErrorMessage)
		http.Error(w, "Unable to talk to Wechat", http.StatusUnprocessableEntity)
		return
	}

	io.WriteString(w, "OK")
}
