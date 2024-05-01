#!/bin/bash
set -m

# Execute /app/DuckChat in the background
/app/DuckChat &

# Add a cron job to restart warp-svc every hour
echo "0 */2 * * * /app/restart_warp.sh" | crontab -
cron &

#
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