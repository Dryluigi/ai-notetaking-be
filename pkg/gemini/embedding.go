package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type EmbedContentRequest struct {
	Model    string  `json:"model"`
	Content  Content `json:"content"`
	TaskType string  `json:"taskType"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type EmbedContentResponse struct {
	Embedding Embedding `json:"embedding"`
}

type Embedding struct {
	Values []float32 `json:"values"`
}

func GetEmbedding(apiKey string, text string, taskType string) (*EmbedContentResponse, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-exp-03-07:embedContent"

	reqBody := EmbedContentRequest{
		Model: "models/gemini-embedding-exp-03-07",
		Content: Content{
			Parts: []Part{
				{Text: text},
			},
		},
		TaskType: taskType,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &EmbedError{Type: ErrTypeMarshalRequest, Message: "Failed to marshal request", Err: err}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &EmbedError{Type: ErrTypeRequestFailed, Message: "Failed to create request", Err: err}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &EmbedError{Type: ErrTypeRequestFailed, Message: "Request failed", Err: err}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &EmbedError{Type: ErrTypeRequestFailed, Message: "Failed to read response", Err: err}
	}

	if resp.StatusCode != http.StatusOK {
		errorType := ErrTypeHTTPStatus
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			errorType = ErrTypeInvalidAPIKey
		}
		return nil, &EmbedError{
			Type:    errorType,
			Message: fmt.Sprintf("API returned status %d", resp.StatusCode),
			Err:     fmt.Errorf("response: %s", string(body)),
		}
	}

	var result EmbedContentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &EmbedError{Type: ErrTypeJSONUnmarshal, Message: "Failed to unmarshal response", Err: err}
	}

	return &result, nil
}
