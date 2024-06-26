package openai_proxy

import (
	"context"
	dgctx "github.com/darwinOrg/go-common/context"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	dglogger "github.com/darwinOrg/go-logger"
	"github.com/darwinOrg/go-web/wrapper"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"time"
)

var DefaultClient *openai.Client

const BizTypeKey = "openaiBizType"
const BizIdKey = "openaiBizId"

type ChatCompletionResponseCallback func(ctx *dgctx.DgContext, request openai.ChatCompletionRequest, response openai.ChatCompletionResponse)

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

func SimpleChatCompletionDefault(ctx *dgctx.DgContext, request openai.ChatCompletionRequest, responseCallback ChatCompletionResponseCallback) (string, error) {
	return SimpleChatCompletion(DefaultClient, ctx, request, responseCallback)
}

func SimpleChatCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.ChatCompletionRequest, responseCallback ChatCompletionResponseCallback) (string, error) {
	response, err := CreateChatCompletion(client, ctx, request, responseCallback)

	if err != nil {
		return "", err
	}

	if response.Usage.PromptTokens == 0 || len(response.Choices) == 0 {
		return "", dgerr.ILLEGAL_OPERATION
	}

	return response.Choices[0].Message.Content, nil
}

func CreateChatCompletionDefault(ctx *dgctx.DgContext, request openai.ChatCompletionRequest, responseCallback ChatCompletionResponseCallback) (openai.ChatCompletionResponse, error) {
	return CreateChatCompletion(DefaultClient, ctx, request, responseCallback)
}

func CreateChatCompletion(client *openai.Client, ctx *dgctx.DgContext, request openai.ChatCompletionRequest, responseCallback ChatCompletionResponseCallback) (openai.ChatCompletionResponse, error) {
	start := time.Now().UnixMilli()
	response, err := client.CreateChatCompletion(context.Background(), request)
	dglogger.Infof(ctx, "[bizType: %s, bizId: %s] create chat completion, request: %+v, response: %+v, error: %v, cost: %d ms",
		GetBizType(ctx), GetBizId(ctx), request, response, err, time.Now().UnixMilli()-start)
	if err == nil {
		dglogger.Infof(ctx, "[bizType: %s, bizId: %s] Model: %s, PromptTokens: %d, CompletionTokens: %d, TotalTokens: %d",
			GetBizType(ctx), GetBizId(ctx), request.Model, response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)

		if responseCallback != nil {
			responseCallback(ctx, request, response)
		}
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
			bizType := c.Query(BizTypeKey)
			if bizType != "" {
				SetBizType(ctx, bizType)
			}

			bizId := c.Query(BizIdKey)
			if bizId != "" {
				SetBizId(ctx, bizId)
			}

			response, err := CreateChatCompletion(client, ctx, *request, nil)
			if err != nil {
				return openai.ChatCompletionResponse{}
			}

			return response
		},
	})
}

func SetBizType(ctx *dgctx.DgContext, bizType string) {
	ctx.SetExtraKeyValue(BizTypeKey, bizType)
}

func GetBizType(ctx *dgctx.DgContext) string {
	bizType := ctx.GetExtraValue(BizTypeKey)
	if bizType == nil {
		return ""
	}

	return bizType.(string)
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
