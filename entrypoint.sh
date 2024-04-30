#!/bin/bash
set -m

# Execute /app/DuckChat in the background
/app/DuckChat &

nohup /usr/local/bin/warp-svc > /app/warp.log &
sleep 2

warp-cli --accept-tos registration new
warp-cli --accept-tos mode proxy

if [[ -n $WARP_LICENSE ]]; then
  warp-cli --accept-tos set-license "${WARP_LICENSE}"
fi

warp-cli --accept-tos connect
fg %1

echo "starting..."