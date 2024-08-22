package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2/google"
)

type NotebookClient struct {
	url   string
	token string
}

type ResponseError struct {
	Code    int
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("Error response status (%d): %s", e.Code, e.Message)
}

const (
	scopes          = "https://www.googleapis.com/auth/cloud-platform"
	serviceEndpoint = "https://australia-southeast1-aiplatform.googleapis.com/v1beta1"
)

func NewNotebookClient(projectID string, location string) (*NotebookClient, error) {

	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, scopes)
	if err != nil {
		return nil, err
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	return &NotebookClient{
		url:   fmt.Sprintf("%s/projects/%s/locations/%s/notebookRuntimeTemplates", serviceEndpoint, projectID, location),
		token: token.AccessToken,
	}, nil
}

func (n *NotebookClient) GetNotebooks() (*ListNotebookRuntimeTemplatesResult, error) {

	body, err := n.curl(http.MethodGet, n.url, nil)

	if err != nil {
		return nil, err
	}

	var templates ListNotebookRuntimeTemplatesResult
	err = json.Unmarshal(body, &templates)

	if err != nil {
		return nil, err
	}
	return &templates, nil
}

func (n *NotebookClient) GetNotebook(name string) (*NotebookRuntimeTemplate, error) {

	notebooks, err := n.GetNotebooks()

	if err != nil {
		return nil, err
	}

	for _, notebook := range notebooks.NotebookRuntimeTemplates {
		if name == *notebook.Name {
			return &notebook, nil
		}
	}
	return nil, fmt.Errorf("could not retrieve notebook: %s", name)
}

func (n *NotebookClient) CreateNotebook(template *NotebookRuntimeTemplate) (*NotebookRuntimeTemplate, error) {

	payload, err := json.Marshal(template)

	if err != nil {
		return nil, err
	}
	body, err := n.curl(http.MethodPost, n.url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	var respJSON map[string]interface{}

	json.Unmarshal(body, &respJSON)

	if err != nil {
		return nil, err
	}

	response, ok := respJSON["response"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not retrieve Response of newly created template")
	}

	name, ok := response["name"].(string)
	if !ok {
		return nil, fmt.Errorf("could not retrieve Name of newly created template")
	}

	return n.GetNotebook(name)
}

func (n *NotebookClient) DeleteNotebookRuntimeTemplate(name string) error {

	url := fmt.Sprintf("%s/%s", serviceEndpoint, name)
	_, err := n.curl(http.MethodDelete, url, nil)
	return err
}

func (nc *NotebookClient) curl(method string, url string, payload io.Reader) ([]byte, error) {

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", nc.token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, &ResponseError{Code: resp.StatusCode, Message: string(body)}
	}

	return body, nil
}

func (n *NotebookRuntimeTemplate) AsString() (string, error) {

	jsonData, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	// Convert the JSON byte slice to a string
	return string(jsonData), nil
}

func (nc *NotebookClient) UpdateNotebook(template *NotebookRuntimeTemplate) error {

	url := fmt.Sprintf("%s/%s?updateMask=encryptionSpec.kmsKeyName", serviceEndpoint, *template.Name)
	payload, err := json.Marshal(template)

	if err != nil {
		return err
	}
	_, err = nc.curl(http.MethodPatch, url, bytes.NewBuffer(payload))

	return err
}
