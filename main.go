package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/liyongxin/prometheus-webhook-wechat/template"
	"github.com/liyongxin/prometheus-webhook-wechat/webrouter"
	"github.com/liyongxin/prometheus-webhook-wechat/chilog"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress    = kingpin.Flag("web.listen-address", "The address to listen on for web interface.").Default(":8765").String()
	corpId = kingpin.Flag("corp.id", "corp id created by wechat app.").Required().String()
	corpSecret = kingpin.Flag("corp.secret", "corp app secret created on wechat.").Required().String()
	corpChatId = kingpin.Flag("corp.chatid", "chatid id created by wechat app.").Required().String()
	requestTimeout   = kingpin.Flag("ding.timeout", "Timeout for invoking Wechat webhook.").Default("5s").Duration()
	templateFileName = kingpin.Flag("template.file", "Customized template file (see template/default.tmpl for example)").Default("").String()
)

func main() {
	allowedLevel := promlog.AllowedLevel{}
	flag.AddFlags(kingpin.CommandLine, &allowedLevel)
	kingpin.Version(version.Print("prometheus-webhook-wechat"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(allowedLevel)
	level.Info(logger).Log("msg", "Starting prometheus-webhook-wechat", "version", version.Info())

	// Load & validate customized template file
	if *templateFileName != "" {
		l := log.With(logger, "filename", *templateFileName)

		b, err := ioutil.ReadFile(*templateFileName)
		if err != nil {
			level.Error(l).Log("msg", "Error reading customizable template file", "err", err)
			os.Exit(1)
		}

		_, err = template.UpdateTemplate(string(b))
		if err != nil {
			level.Error(l).Log("msg", "Error parsing template file", "err", err)
			os.Exit(1)
		}

		level.Info(l).Log("msg", "Using customized template")
	} else {
		level.Info(logger).Log("msg", "Using default template")
	}

	// Print current profile configuration
	level.Info(logger).Log("msg", fmt.Sprintf("send message to following chatgroup id: %s", corpChatId))

	r := chi.NewRouter()
	//r.Use(middleware.CloseNotify)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&chilog.KitLogger{Logger: logger}))
	r.Use(middleware.Recoverer)

	weChatResource := &webrouter.WechatResource{
		Logger:   logger,
		CorpId: *corpId,
		WechatUrl: "https://qyapi.weixin.qq.com/cgi-bin/",
		CorpSecret: *corpSecret,
		CorpChatId: *corpChatId,
		HttpClient: &http.Client{
			Timeout: *requestTimeout,
			Transport: &http.Transport{
				Proxy:             http.ProxyFromEnvironment,
				DisableKeepAlives: true,
			},
		},
	}
	r.Mount("/v1/wechat", weChatResource.Routes())

	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, r); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
