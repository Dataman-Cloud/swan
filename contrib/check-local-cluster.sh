#!/bin/bash

function check(){
	local n=0
	local cnames=( "swan_swan-master_1"  "swan_swan-agent_1"  )

	for cname in `echo ${cnames[*]}`
	do
		status=$(docker inspect -f "{{.State.Health.Status}}" $cname)
		echo "$cname --> $status"

		if [ "${status}" == "healthy" ]; then
			((n++))
		fi
	done

	if [ $n -eq 2 ]; then
		return 0
	fi
	return 1
}

echo "waitting for local cluster is ready ...."

for i in `seq 600`
do
	sleep 1
	if check; then
		echo "local cluster ready!"
		exit 0
	else
		echo "local cluster not ready, waitting ..."
	fi
done

echo "local cluster not ready"
exit 1
