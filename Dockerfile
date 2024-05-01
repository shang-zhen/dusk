# First stage: Ubuntu setup
FROM ubuntu:22.04 AS buildwarp

RUN apt update && apt install -y curl gpg lsb-release wget

RUN curl -fsSL https://pkg.cloudflareclient.com/pubkey.gpg | gpg --yes --dearmor --output /usr/share/keyrings/cloudflare-warp-archive-keyring.gpg
RUN echo "deb [signed-by=/usr/share/keyrings/cloudflare-warp-archive-keyring.gpg] https://pkg.cloudflareclient.com/ $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/cloudflare-client.list
RUN apt update && apt install -y cloudflare-warp

# Second stage: Go build
FROM golang:1.22 AS buildgo

WORKDIR /go/src/duck
COPY main.go ./
COPY go.mod ./
COPY go.sum ./
RUN go get -d -v ./
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o DuckChat .

# Final stage: Alpine setup
FROM ubuntu:22.04
RUN apt update && \
    apt-get install -y libdbus-1-3 curl cron

COPY --from=buildwarp /usr/bin/warp-cli /usr/bin/warp-svc /usr/local/bin/

#duck
COPY --from=buildgo /go/src/duck/DuckChat /app/DuckChat
RUN chmod +x /app/DuckChat
#restart
COPY restart_warp.sh /app/restart_warp.sh
RUN chmod +x /app/restart_warp.sh
#entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
