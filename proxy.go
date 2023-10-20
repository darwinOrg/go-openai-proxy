package openai_proxy

import (
	"context"
	dgctx "github.com/darwinOrg/go-common/context"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-openai"
	"github.com/darwinOrg/go-web/wrapper"
	"github.com/gin-gonic/gin"
	"time"
)

// NewProxyClient creates new OpenAI API proxy client.
func NewProxyClient(proxyBaseUrl string) *openai.Client {
	return NewProxyClientWithToken(proxyBaseUrl, "none")
}

// NewProxyClientWithToken creates new OpenAI API proxy client with auth token.
func NewProxyClientWithToken(proxyBaseUrl string, authToken string) *openai.Client {
	config := openai.DefaultConfig(authToken)
	config.BaseURL = proxyBaseUrl
	return openai.NewClientWithConfig(config)
}

func CreateChatCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	start := time.Now().UnixMilli()
	response, err := client.CreateChatCompletion(context.Background(), request)
	dglogger.Infof(ctx, "create chat completion, request: %+v, response: %+v, error: %v, cost: %d ms",
		request, response, err, time.Now().UnixMilli()-start)
	return response, err
}

func CreateCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.CompletionRequest) (openai.CompletionResponse, error) {
	start := time.Now().UnixMilli()
	response, err := client.CreateCompletion(context.Background(), request)
	dglogger.Infof(ctx, "create completion, request: %+v, response: %+v, error: %v, cost: %d ms",
		request, response, err, time.Now().UnixMilli()-start)
	return response, err
}

func BindRouter(rg *gin.RouterGroup, client *openai.Client) {
	wrapper.Post(&wrapper.RequestHolder[openai.ChatCompletionRequest, openai.ChatCompletionResponse]{
		RouterGroup:  rg,
		RelativePath: "/chat/completions",
		NonLogin:     true,
		BizHandler: func(_ *gin.Context, ctx *dgctx.DgContext, request *openai.ChatCompletionRequest) openai.ChatCompletionResponse {
			response, err := CreateChatCompletion(client, ctx, *request)
			if err != nil {
				return openai.ChatCompletionResponse{}
			}

			return response
		},
	})

	wrapper.Post(&wrapper.RequestHolder[openai.CompletionRequest, openai.CompletionResponse]{
		RouterGroup:  rg,
		RelativePath: "/completions",
		NonLogin:     true,
		BizHandler: func(_ *gin.Context, ctx *dgctx.DgContext, request *openai.CompletionRequest) openai.CompletionResponse {
			response, err := CreateCompletion(client, ctx, *request)
			if err != nil {
				return openai.CompletionResponse{}
			}

			return response
		},
	})

}
