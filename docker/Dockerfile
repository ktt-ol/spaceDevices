FROM golang:1.13

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/0.5.2/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN mkdir -p /go/src/github.com/ktt-ol/spaceDevices

ENV HOME=/tmp
WORKDIR /go/src/github.com/ktt-ol/spaceDevices
