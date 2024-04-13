package yandexgpt_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	yandexgpt "github.com/playmixer/yandex/GPT"
)

func init() {
	godotenv.Load()
}

func TestNewEmptyApiKey(t *testing.T) {
	_, err := yandexgpt.New("", "")
	if err != nil {
		return
	}
	t.Fatal("FAILED", err)
}

func TestNewEmptyFolderId(t *testing.T) {
	_, err := yandexgpt.New("qw", "")
	if err != nil {
		return
	}
	t.Fatal("FAILED", err)
}

func TestSendMessageHi(t *testing.T) {
	gpt, err := yandexgpt.New(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER"))
	if err != nil {
		t.Fatal(err)
	}
	req := gpt.NewRequest()
	req.AddMessage(yandexgpt.YandexGPTRequestMessages{Role: yandexgpt.GPTRoleUser, Text: "Привет"})
	resp, err := req.Do()
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Println(req)
	if resp.StatusCode != 200 {
		t.Fatalf("response not OK status=%v", resp.StatusCode)
	}

}

func TestSendMessageStream(t *testing.T) {
	gpt, err := yandexgpt.New(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER"))
	if err != nil {
		t.Fatal(err)
	}
	req := gpt.NewRequest()
	req.AddMessage(yandexgpt.YandexGPTRequestMessages{Role: yandexgpt.GPTRoleUser, Text: "как ты можешь помочь как голосовой ассистент?"})
	req.CompletionOptions.Stream = true
	ch := make(chan yandexgpt.YandexGPTResponse, 2)
	err = req.DoStream(ch)
	if err != nil {
		t.Fatal(err)
	}
	for resp := range ch {
		fmt.Println(resp)
		if resp.StatusCode != 200 {
			t.Fatalf("response not OK status=%v", resp.StatusCode)
		}
	}

}
