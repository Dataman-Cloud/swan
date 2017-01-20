#!/bin/bash


http --check-status --ignore-stdin --timeout=4.5 put $SERVER_PATH/$PATH_PREFIX/apps/$APP_NAME @template-replicates-versioned.json &>/dev/null

wait_a_moment # wait scalling up operation done

task0Version=`http get http://localhost:9999/v_beta/apps/$APP_NAME  | jq '.tasks | .[] | select( .id | contains("0-")) .versionId'`
task1Version=`http get http://localhost:9999/v_beta/apps/$APP_NAME  | jq '.tasks | .[] | select( .id | contains("1-")) .versionId'`
appVersion=`http get http://localhost:9999/v_beta/apps/$APP_NAME  | jq '.currentVersion.id'`

if [ "$task0Version" != "$appVersion" ]
then
  ok "app $APP_NAME successfully rolling updated with slot 0 have version $task0Version and app have version $appVersion"
else
  fail "app $APP_NAME failed rolling updated with slot 0 have version $task0Version and app have version $appVersion"
fi
