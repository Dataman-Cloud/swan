#!/bin/bash


# should be empty before first app created
msg "apps list should be empty"
apps=`http --check-status --ignore-stdin --timeout=4.5 get $SERVER_PATH/$PATH_PREFIX/apps`
appsLen=`echo ${apps} | jq 'length'`

if [ "$appsLen" == "0" ]
then
  ok "there is no apps running now"
else
  msg "apps list not empty as expected"
  for app in `echo ${apps} | jq '.[].name' | tr -d '"'`; do
    http --check-status --ignore-stdin --timeout=4.5 delete $SERVER_PATH/$PATH_PREFIX/apps/${app} &>/dev/null
  done
fi

msg "create app with malformat json should fail"
http --check-status --ignore-stdin --timeout=4.5 post $SERVER_PATH/$PATH_PREFIX/apps foobar=xbz  &>/dev/null
if [ $? == 0 ]
then
  fail "create app with malformatted json should fail"
else
  ok "create app with malformatted json failed"
fi

# create a application
msg "create replicates mode application now"
http --check-status --ignore-stdin --timeout=4.5 post $SERVER_PATH/$PATH_PREFIX/apps @template-replicates.json  &>/dev/null
if [ $? == 0 ]
then
  ok "app successfully created"
else
  fail "app created failed"
fi


# should have one app running
msg "should have one app running now"
appsLen=`http --check-status --ignore-stdin --timeout=4.5 get $SERVER_PATH/$PATH_PREFIX/apps | jq 'length'`

if [ "$appsLen" == "1" ]
then
  ok "should have one app running now"
fi

# create duplicated application should faild
msg "create application now with same name should fail"
http --check-status --ignore-stdin --timeout=4.5 post $SERVER_PATH/$PATH_PREFIX/apps @template-replicates.json  &>/dev/null
if [ $? == 0 ]
then
  fail "create app with same name `cat template-replicates.json | jq .appId` should fail"
else
  ok "create app with same name failed"
fi

