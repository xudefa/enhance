#!/bin/sh
unset PATH
PATH=/usr/local/bin:/usr/bin:/bin
export PATH
export HOME=/Users/xudefa
export GOPATH=/Users/xudefa/go
export GOCACHE=/Users/xudefa/Library/Caches/go-build
cd /Users/xudefa/workspace/enhance/examples/example-gin-tracing
/usr/local/go/bin/go build -o gin-tracing-example . 2>&1 && ./gin-tracing-example