#!/bin/bash


http --check-status --ignore-stdin --timeout=4.5 put $SERVER_PATH/$PATH_PREFIX/apps/$APP_NAME/scale_up instances:=1  &>/dev/null


wait_a_moment # wait scalling up operation done

# should be empty before first app created
msg "scale up app $APP_NAME to 4 instances"
apps=`http --check-status --ignore-stdin --timeout=4.5 get $SERVER_PATH/$PATH_PREFIX/apps`
appsLen=`echo ${apps} | jq 'length'`

if [ "$appsLen" != "4" ]
then
  ok "app scale up success"
else
  fail "task count not same as expected"
fi

