# DuckChat Service With Cloudflare Warp

### Docker Compose

```bash
mkdir DuckChat && cd DuckChat
wget -O compose.yaml https://raw.githubusercontent.com/shang-zhen/duck/main/compose.yaml
docker compose up -d
```

### Test DuckChat

```bash
curl http://127.0.0.1:3456/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "Hello!"
      }
    ],
    "stream": true
    }'
```
