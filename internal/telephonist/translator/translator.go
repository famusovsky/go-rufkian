package translator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"resty.dev/v3"
)

type yaClient struct {
	folderID string
	client   *resty.Client
}

func NewYaClient(folderID, key string) IClient {
	return &yaClient{
		folderID: folderID,
		client: resty.New().
			SetDisableWarn(true).
			SetBaseURL(yaTranslateURL).
			SetTimeout(30*time.Second).
			SetHeader("Authorization", fmt.Sprintf("Api-Key %s", key)),
	}
}

type yaRequest struct {
	Texts              []string `json:"texts"`
	SourceLanguageCode string   `json:"sourceLanguageCode"`
	TargetLanguageCode string   `json:"targetLanguageCode"`
	FolderID           string   `json:"folderId"`
}

type yaResponseTranslation struct {
	Text                 string `json:"text"`
	DetectedLanguageCode string `json:"detectedLanguageCode"`
}

type yaResponse struct {
	Translations []yaResponseTranslation `json:"translations"`
}

func (c *yaClient) Translate(texts []string) ([]string, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty texts")
	}

	payload := yaRequest{
		Texts:              texts,
		FolderID:           c.folderID,
		SourceLanguageCode: german,
		TargetLanguageCode: russian,
	}

	req := c.client.R().
		SetContentType("application/json").
		SetBody(payload).
		SetMethod(resty.MethodPost)

	rawResponse, err := req.Send()
	if err != nil {
		return nil, err
	}

	if rawResponse.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response from translator with status: %s", rawResponse.Status())
	}

	var response yaResponse
	if err := json.Unmarshal(rawResponse.Bytes(), &response); err != nil {
		return nil, err
	}

	ret := make([]string, 0, len(response.Translations))
	for _, t := range response.Translations {
		ret = append(ret, t.Text)
	}

	return ret, nil
}
