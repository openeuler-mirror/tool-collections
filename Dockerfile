FROM golang:stretch

COPY . /root/tool-collections

RUN cd /root/tool-collections && \
GOFLAGS=-mod=vendor go build -o validator ./cmd/tools && \
cp validator /usr/bin

