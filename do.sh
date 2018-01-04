#!/bin/bash

usage() {
    echo "Usage $0 (build-linux|sync)"
}

if [ "$1" == "" ]; then
    usage
    exit 1
fi


while (( "$#" )); do
    case "$1" in
        build-linux)
            env GOOS=linux GOARCH=amd64 go build cmd/spaceDevices.go
            ;;
        test-sync)
            rsync -n -avzi --delete spaceDevices webUI root@spacegate:/home/status/spaceDevices2/
            ;;
        sync)
            rsync -avzi --delete spaceDevices webUI root@spacegate:/home/status/spaceDevices2/
            ;;
        *)
            usage
            exit 1
    esac
    shift
done