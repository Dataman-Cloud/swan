#!/bin/bash

#build app
msg "build app"
http --check-status --ignore-stdin --timeout=4.5 post $SERVER_PATH/v1/apps @../example/template.json | jq .data | grep "success" 1>/dev/null 2>&1
if [ "$?" != "0" ]
then
	fail "build app error"
else
	ok "successfully build a app"
fi

#list apps
msg "list apps"
assert_status_code "get" "v1/apps" 200

#get app
msg "get app"
assert_status_code "get" "v1/apps/1" 200

#update app
msg "update app"

#scale app






