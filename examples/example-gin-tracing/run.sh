#!/usr/bin/env bash
# 清理所有环境变量
exec env -i \
  PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin \
  HOME="$HOME" \
  USER="$USER" \
  LOGNAME="$LOGNAME" \
  TMPDIR=/tmp \
  GOPATH="$HOME/go" \
  GOCACHE="$HOME/Library/Caches/go-build" \
  GOFLAGS="-mod=mod" \
  /bin/bash -c 'cd /Users/xudefa/workspace/enhance/examples/example-gin-tracing && go run main.go 2>&1'