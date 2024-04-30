package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/proxy"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream bool `json:"stream"`
}

type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}

type OpenAIChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason *string     `json:"finish_reason"`
}

type OpenAIDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type OpenAINonStreamResponse struct {
	ID      string                  `json:"id"`
	Object  string                  `json:"object"`
	Created int64                   `json:"created"`
	Model   string                  `json:"model"`
	Choices []OpenAINonStreamChoice `json:"choices"`
}

type OpenAINonStreamChoice struct {
	Index        int         `json:"index"`
	Message      OpenAIDelta `json:"message"`
	FinishReason *string     `json:"finish_reason"`
}

type DuckDuckGoResponse struct {
	Role    string `json:"role"`
	Message string `json:"message"`
	Created int64  `json:"created"`
	ID      string `json:"id"`
	Action  string `json:"action"`
	Model   string `json:"model"`
}

func getRandomUserAgent() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	userAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0  uacq",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Iron Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; rv:111.0) Gecko/20100101 Firefox/111.0",
	}
	return userAgents[r.Intn(len(userAgents))]
}

func chatWithDuckDuckGo(c *gin.Context, messages []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}, stream bool) {
	userAgent := getRandomUserAgent()

	headers := map[string]string{
		"User-Agent":      userAgent,
		"Accept":          "text/event-stream",
		"Accept-Language": "de,en-US;q=0.7,en;q=0.3",
		"Accept-Encoding": "gzip, deflate, br",
		"Referer":         "https://duckduckgo.com/",
		"Content-Type":    "application/json",
		"Origin":          "https://duckduckgo.com",
		"Connection":      "keep-alive",
		"Cookie":          "dcm=1",
		"Sec-Fetch-Dest":  "empty",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "same-origin",
		"Pragma":          "no-cache",
	}

	statusURL := "https://duckduckgo.com/duckchat/v1/status"
	chatURL := "https://duckduckgo.com/duckchat/v1/chat"
	socks5ProxyURL := "0.0.0.0:40000"
	auth := &proxy.Auth{}
	dialer, err := proxy.SOCKS5("tcp", socks5ProxyURL, auth, proxy.Direct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
	}

	// get vqd_4
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("x-vqd-accept", "1")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	vqd4 := resp.Header.Get("x-vqd-4")

	payload := map[string]interface{}{
		"model":    "gpt-3.5-turbo-0125",
		"messages": messages,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dialer, err = proxy.SOCKS5("tcp", socks5ProxyURL, auth, proxy.Direct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transport = &http.Transport{
		Dial: dialer.Dial,
	}

	client = &http.Client{
		Transport: transport,
	}

	// 创建新的请求
	req, err = http.NewRequest("POST", chatURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置请求头
	req.Header.Set("x-vqd-4", vqd4)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 使用带有代理设置的客户端执行请求
	resp, err = client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	flusher, _ := c.Writer.(http.Flusher)

	var response OpenAIResponse
	var nonStreamResponse OpenAINonStreamResponse
	response.Choices = make([]OpenAIChoice, 1)
	nonStreamResponse.Choices = make([]OpenAINonStreamChoice, 1)

	var responseContent string

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if bytes.HasPrefix(line, []byte("data: ")) {
			chunk := line[6:]

			if bytes.HasPrefix(chunk, []byte("[DONE]")) {
				if !stream {
					nonStreamResponse.Choices[0].Message.Content = responseContent
					nonStreamResponse.Choices[0].Message.Role = "assistant"
					nonStreamResponse.Choices[0].FinishReason = new(string)
					*nonStreamResponse.Choices[0].FinishReason = "stop"
					c.JSON(http.StatusOK, nonStreamResponse)
					return
				} else {
					stopData := OpenAIResponse{
						ID:      "chatcmpl-9HOzx2PhnYCLPxQ3Dpa2OKoqR2lgl",
						Object:  "chat.completion",
						Created: 1713934697,
						Model:   "gpt-3.5-turbo-0125",
						Choices: []OpenAIChoice{
							{
								Index:        0,
								FinishReason: stringPtr("stop"),
							},
						},
					}
					stopDataBytes, _ := json.Marshal(stopData)
					c.Data(http.StatusOK, "application/json", []byte("data: "))
					c.Data(http.StatusOK, "application/json", stopDataBytes)
					c.Data(http.StatusOK, "application/json", []byte("\n\n"))
					flusher.Flush()

					c.Data(http.StatusOK, "application/json", []byte("data: [DONE]\n\n"))
					flusher.Flush()
					return
				}
			}

			var data DuckDuckGoResponse
			decoder := json.NewDecoder(bytes.NewReader(chunk))
			decoder.UseNumber()
			err = decoder.Decode(&data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			response.ID = data.ID
			response.Object = "chat.completion"
			response.Created = data.Created
			response.Model = data.Model
			nonStreamResponse.ID = data.ID
			nonStreamResponse.Object = "chat.completion"
			nonStreamResponse.Created = data.Created
			nonStreamResponse.Model = data.Model
			responseContent += data.Message

			if stream {
				response.Choices[0].Delta.Content = data.Message

				responseBytes, err := json.Marshal(response)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.Data(http.StatusOK, "application/json", append(append([]byte("data: "), responseBytes...), []byte("\n\n")...))
				flusher.Flush()

				response.Choices[0].Delta.Content = ""
			}

		}
	}
}

func stringPtr(s string) *string {
	return &s
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello world.",
		})
	})

	r.POST("/v1/chat/completions", func(c *gin.Context) {
		// 获取 Authorization 请求头
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader != "Bearer 000123123" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid bearer token"})
			return
		}

		var req OpenAIRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// only support user role
		for i := range req.Messages {
			if req.Messages[i].Role == "system" {
				req.Messages[i].Role = "user"
			}
		}
		// set model to gpt-3.5-turbo-0125
		req.Model = "gpt-3.5-turbo-0125"
		chatWithDuckDuckGo(c, req.Messages, req.Stream)
	})

	r.GET("/v1/models", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"object": "list",
			"data": []gin.H{
				{
					"id":       "gpt-3.5-turbo-0125",
					"object":   "model",
					"created":  1692901427,
					"owned_by": "system",
				},
			},
		})
	})

	r.Run(":3456")

	//显示已经启动
	println("Server started on :3456")

}
