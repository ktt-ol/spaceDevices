#!/bin/bash

usage() {
    echo "Usage $0 (build-linux|docker-image|docker-build|test-sync|sync)"
}

if [ "$1" == "" ]; then
    usage
    exit 1
fi


while (( "$#" )); do
    case "$1" in
        build-linux)
            env GOOS=linux GOARCH=amd64 go build cmd/spaceDevices/spaceDevices.go
            env GOOS=linux GOARCH=amd64 go build cmd/unkownDevices/listUnkown.go
            ;;
        docker-image)
            docker build -t space-devices-build docker/
            ;;
        docker-build)
            docker run --rm -it -v $(pwd):/go/src/spaceDevices space-devices-build dep ensure -v -vendor-only && ./do.sh build-linux
            ;;
        test-sync)
            rsync -n -avzi --delete spaceDevices listUnkown webUI macVendorDb.csv root@spacegate:/home/status/spaceDevices2/
            ;;
        sync)
            rsync -avzi --delete spaceDevices listUnkown webUI macVendorDb.csv root@spacegate:/home/status/spaceDevices2/
            ;;
        *)
            usage
            exit 1
    esac
    shift
done