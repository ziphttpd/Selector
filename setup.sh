#!/bin/bash
SCRIPT=$0
TARGET_DIR=$1

if [ -d "${TARGET_DIR}" ]; then
	echo "Usage: ${SCRIPT} {install-directory}"
	exit 1
fi

SCRIPT_DIR=$(cd $(dirname ${SCRIPT}); pwd)
cd ${SCRIPT_DIR}
go run github.com/rakyll/statik -f -src=static
go build -o ${TARGET_DIR}/selector main.go

exit 0
