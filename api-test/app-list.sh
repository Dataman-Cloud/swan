#!/bin/bash


# list applications
msg "list applications"
assert_status_code "GET" "${PATH_PREFIX}/apps" 200

