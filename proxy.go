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

const BizIdKey = "openaiBizId"

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

	if response.Usage.PromptTokens == 0 || len(response.Choices) == 0 {
		return "", dgerr.ILLEGAL_OPERATION
	}

	return response.Choices[0].Message.Content, nil
}

func CreateChatCompletionDefault(ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return CreateChatCompletion(DefaultClient, ctx, request)
}

func CreateChatCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	start := time.Now().UnixMilli()
	response, err := client.CreateChatCompletion(context.Background(), request)
	dglogger.Infof(ctx, "[bizId: %s] create chat completion, request: %+v, response: %+v, error: %v, cost: %d ms",
		GetBizId(ctx), request, response, err, time.Now().UnixMilli()-start)
	if err == nil {
		dglogger.Infof(ctx, "[bizId: %s] Model: %s, PromptTokens: %d, CompletionTokens: %d, TotalTokens: %d",
			GetBizId(ctx), request.Model, response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)
	}

	return response, err
}

func BindRouterDefault(rg *gin.RouterGroup) {
	BindRouter(rg, DefaultClient)
}

func BindRouter(rg *gin.RouterGroup, client *openai.Client) {
	wrapper.Post(&wrapper.RequestHolder[openai.ChatCompletionRequest, openai.ChatCompletionResponse]{
		RouterGroup:  rg,
		RelativePath: "/chat/completions",
		NonLogin:     true,
		BizHandler: func(c *gin.Context, ctx *dgctx.DgContext, request *openai.ChatCompletionRequest) openai.ChatCompletionResponse {
			ctx.SetExtraKeyValue(BizIdKey, c.Query(BizIdKey))
			response, err := CreateChatCompletion(client, ctx, *request)
			if err != nil {
				return openai.ChatCompletionResponse{}
			}

			return response
		},
	})
}

func SetBizId(ctx *dgctx.DgContext, bizId string) {
	ctx.SetExtraKeyValue(BizIdKey, bizId)
}

func GetBizId(ctx *dgctx.DgContext) string {
	bizId := ctx.GetExtraValue(BizIdKey)
	if bizId == nil {
		return ""
	}

	return bizId.(string)
}
