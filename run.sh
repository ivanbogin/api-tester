#!/bin/bash
(export PORT=8080; export REDIS_ADDR=127.0.0.1:6379; go run main.go)
