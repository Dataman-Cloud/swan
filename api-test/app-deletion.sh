#!/bin/bash


http --check-status --ignore-stdin --timeout=4.5 delete $SERVER_PATH/$PATH_PREFIX/apps/$APP_NAME &>/dev/null

wait_a_moment # wait scalling up operation done

# should be empty before first app created
msg "app list should be empty"
apps=`http --check-status --ignore-stdin --timeout=4.5 get $SERVER_PATH/$PATH_PREFIX/apps`
appsLen=`echo ${apps} | jq 'length'`

if [ "$appsLen" == "0" ]
then
  ok "app $APP_NAME remove successfully"
else
  fail "app deletion failed"
fi

