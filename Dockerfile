FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Yongxin Li <yxli@alauda.io>

COPY prometheus-webhook-wechat  /bin/prometheus-webhook-wechat
COPY template/default.tmpl        /usr/share/prometheus-webhook-dingtalk/template/default.tmpl
COPY run.sh  /

EXPOSE      8765

CMD /bin/sh /run.sh
