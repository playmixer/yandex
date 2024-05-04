package yandexgpt

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	URLGPTCompletion string = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
)

type YandexGPT struct {
	apiKey   string
	folderId string
}

type GPTRole string
type GPTTypeModelURI string

const (
	GPTRoleSystem    GPTRole = "system"
	GPTRoleAssistent GPTRole = "assistent"
	GPTRoleUser      GPTRole = "user"

	GPTTypeModelURIPro        GPTTypeModelURI = "YandexGPTPro"
	GPTTypeModelURILite       GPTTypeModelURI = "YandexGPTLite"
	GPTTypeModelURIShort      GPTTypeModelURI = "YandexGPTShort"
	GPTTypeModelURIDataSphere GPTTypeModelURI = "YandexGPTDataSphere"
)

func New(apikey, folder string) (*YandexGPT, error) {

	if apikey == "" {
		return nil, fmt.Errorf("api key unvalid")
	}
	if folder == "" {
		return nil, fmt.Errorf("folderId unvalid")
	}

	return &YandexGPT{
		apiKey:   apikey,
		folderId: folder,
	}, nil
}

func (g *YandexGPT) GetModelUri(model GPTTypeModelURI) string {
	switch model {
	case GPTTypeModelURIPro:
		return fmt.Sprintf("gpt://%s/yandexgpt/latest", g.folderId)
	case GPTTypeModelURILite:
		return fmt.Sprintf("gpt://%s/yandexgpt-lite/latest", g.folderId)
	case GPTTypeModelURIShort:
		return fmt.Sprintf("gpt://%s/summarization/latest", g.folderId)
	default:
		return fmt.Sprintf("gpt://%s/yandexgpt/latest", g.folderId)
	}
}

type YandexGPTRequestCompletionOptions struct {
	Stream       bool    `json:"stream"`
	Templerature float32 `json:"temperature"`
	MaxTokens    int64   `json:"maxTokens"`
}

type YandexGPTMessage struct {
	Role GPTRole `json:"role"`
	Text string  `json:"text"`
}

type YandexGPTRequest struct {
	ModelURI          string                            `json:"modelUri"`
	CompletionOptions YandexGPTRequestCompletionOptions `json:"completionOptions"`
	Messages          []YandexGPTMessage                `json:"messages"`
	ygpt              YandexGPT
}

func (g *YandexGPT) NewRequest() *YandexGPTRequest {

	return &YandexGPTRequest{
		ModelURI: g.GetModelUri(GPTTypeModelURILite),
		CompletionOptions: YandexGPTRequestCompletionOptions{
			Stream:       false,
			Templerature: 0.3,
			MaxTokens:    100,
		},
		Messages: []YandexGPTMessage{},
		ygpt:     *g,
	}
}

func (req *YandexGPTRequest) AddMessage(message YandexGPTMessage) {
	req.Messages = append(req.Messages, message)
}

func (req *YandexGPTRequest) AddMessages(messages []YandexGPTMessage) {
	req.Messages = append(req.Messages, messages...)
}

/*
Enum representing the generation status of the alternative.

ALTERNATIVE_STATUS_UNSPECIFIED: Unspecified generation status. - ALTERNATIVE_STATUS_PARTIAL: Partially generated alternative.
ALTERNATIVE_STATUS_TRUNCATED_FINAL: Incomplete final alternative resulting from reaching the maximum allowed number of tokens.
ALTERNATIVE_STATUS_FINAL: Final alternative generated without running into any limits.
ALTERNATIVE_STATUS_CONTENT_FILTER: Generation was stopped due to the discovery of potentially sensitive content in the prompt or generated response. To fix, modify the prompt and restart generation.
*/
type YandexGPTResponseAlternativeStatus string

const (
	AlternativeStatusUnspecified    YandexGPTResponseAlternativeStatus = "ALTERNATIVE_STATUS_UNSPECIFIED"
	AlternativeStatusTruncatedFinal YandexGPTResponseAlternativeStatus = "ALTERNATIVE_STATUS_TRUNCATED_FINAL"
	AlternativeStatusFinal          YandexGPTResponseAlternativeStatus = "ALTERNATIVE_STATUS_FINAL"
	AlternativeStatusContentFilter  YandexGPTResponseAlternativeStatus = "ALTERNATIVE_STATUS_CONTENT_FILTER"
)

type YandexGPTResponseAlternative struct {
	Message YandexGPTMessage                   `json:"message"`
	Status  YandexGPTResponseAlternativeStatus `json:"status"`
}

type YandexGPTResponseUsage struct {
	InputTextTokens  string `json:"inputTextTokens"`
	CompletionTokens string `json:"completionTokens"`
	TotalTokens      string `json:"totalTokens"`
}

type YandexGPTResponse struct {
	Result struct {
		Alternatives []YandexGPTResponseAlternative `json:"alternatives"`
		Usage        YandexGPTResponseUsage         `json:"usage"`
		ModelVersion string                         `json:"modelVersion"`
	} `json:"result"`
	Error struct {
		GrpcCode   int64  `json:"grpcCode"`
		HTTPCode   int64  `json:"httpCode"`
		Message    string `json:"message"`
		HTTPStatus string `json:"httpStatus"`
	} `json:"error"`
	StatusCode int
}

func (req *YandexGPTRequest) Do() (*YandexGPTResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", URLGPTCompletion, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", req.ygpt.apiKey))

	client := &http.Client{}
	if os.Getenv("TLS") == "0" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result := &YandexGPTResponse{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.StatusCode = response.StatusCode

	return result, nil
}

func (req *YandexGPTRequest) DoStream(ch chan<- YandexGPTResponse) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", URLGPTCompletion, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", req.ygpt.apiKey))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	go func(resp *http.Response) {
		defer response.Body.Close()
		defer close(ch)
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			result := &YandexGPTResponse{}

			err = json.Unmarshal(scanner.Bytes(), result)
			if err != nil {
				log.Fatal(err)
			}
			result.StatusCode = response.StatusCode
			ch <- *result
		}
	}(response)

	return nil
}

// func Post(url string, headers map[string]string)
