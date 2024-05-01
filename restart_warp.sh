#!/bin/bash
pkill warp-svc
nohup /usr/local/bin/warp-svc > /app/warp.log &