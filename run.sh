#!/usr/bin/env bash

corp_id=$CORPID
corp_secret=$CORPSECRET
chat_id=$CHATID
/bin/prometheus-webhook-wechat --corp.id=$CORPID --corp.secret=$CORPSECRET --corp.chatid=$CHATID