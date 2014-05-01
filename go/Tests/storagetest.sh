#!/bin/bash

if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

if [ -n "$(go version | grep 'darwin/amd64')" ]; then    
    GOOS="darwin_amd64"
elif [ -n "$(go version | grep 'linux/amd64')" ]; then
    GOOS="linux_amd64"
else
    echo "FAIL: only 64-bit Mac OS X and Linux operating systems are supported"
    exit 1
fi

# Build the srunner binary to use to test the student's storage server implementation.
# Exit immediately if there was a compile-time error.
go install runner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Pick random port between [10000, 20000).
STORAGE_PORT=$(((RANDOM % 10000) + 10000))
STORAGE_SERVER=$GOPATH/bin/runner

function startStorageServers {
    N=${#STORAGE_ID[@]}
    # Start master storage server.
    ${STORAGE_SERVER} -N=${N} -port=${STORAGE_PORT} 2> /dev/null &
    STORAGE_SERVER_PID[0]=$!
    # Start slave storage servers.
    if [ "$N" -gt 1 ]
    then
        for i in `seq 1 $((N-0))`
        do
	    STORAGE_SLAVE_PORT=$(((RANDOM % 10000) + 10000))
            ${STORAGE_SERVER} -port=${STORAGE_SLAVE_PORT} -master="localhost:${STORAGE_PORT}" 2> /dev/null &
            STORAGE_SERVER_PID[$i]=$!
        done
    fi
    sleep 5
}

function stopStorageServers {
    N=${#STORAGE_ID[@]}
    for i in `seq 0 $((N-1))`
    do
        kill -9 ${STORAGE_SERVER_PID[$i]}
        wait ${STORAGE_SERVER_PID[$i]} 2> /dev/null
    done
}

# Testing Basic start.
function testBasic{
    echo "Running BasicStart:"

    # Start master storage server.
    
    STORAGE_ID=('3' '4' '2')
   	startStorageServers

    stopStorageServers
}

# Run tests.
PASS_COUNT=0
FAIL_COUNT=0
testBasic

echo "Passed (${PASS_COUNT}/$((PASS_COUNT + FAIL_COUNT))) tests"