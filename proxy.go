package openai_proxy

import (
	"context"
	dgctx "github.com/darwinOrg/go-common/context"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-openai"
	"github.com/darwinOrg/go-web/wrapper"
	"github.com/gin-gonic/gin"
	"time"
)

var DefaultClient *openai.Client

func NewProxyClientDefault(baseUrl string) {
	DefaultClient = NewProxyClient(baseUrl)
}

func NewProxyClient(baseUrl string) *openai.Client {
	return NewProxyClientWithToken(baseUrl, "none")
}

func NewProxyClientWithTokenDefault(baseUrl string, authToken string) {
	DefaultClient = NewProxyClientWithToken(baseUrl, authToken)
}

func NewProxyClientWithToken(baseUrl string, authToken string) *openai.Client {
	config := openai.DefaultConfig(authToken)
	config.BaseURL = baseUrl
	return openai.NewClientWithConfig(config)
}

func SimpleChatCompletionDefault(ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (string, error) {
	return SimpleChatCompletion(DefaultClient, ctx, request)
}

func SimpleChatCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (string, error) {
	response, err := CreateChatCompletion(client, ctx, request)

	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", dgerr.SYSTEM_ERROR
	}

	return response.Choices[0].Message.Content, nil
}

func CreateChatCompletionDefault(ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return CreateChatCompletion(DefaultClient, ctx, request)
}

func CreateChatCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	start := time.Now().UnixMilli()
	response, err := client.CreateChatCompletion(context.Background(), request)
	dglogger.Infof(ctx, "create chat completion, request: %+v, response: %+v, error: %v, cost: %d ms",
		request, response, err, time.Now().UnixMilli()-start)
	return response, err
}

func CreateCompletionDefault(ctx *dgctx.DgContext, request openai.CompletionRequest) (openai.CompletionResponse, error) {
	return CreateCompletion(DefaultClient, ctx, request)
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
