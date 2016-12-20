#!/bin/bash

. ./config_and_precheck.sh
. ./functions.sh

. ./app-list.sh

. ./app-create.sh

wait_a_moment # wait previous operation done

. ./app-scale-up.sh

wait_a_moment # wait previous operation done

. ./app-scale-down.sh



wait_a_moment # wait previous operation done

. ./app-rolling-update.sh


wait_a_moment # wait previous operation done

. ./app-deletion.sh


