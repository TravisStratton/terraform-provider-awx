package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type AwxClient struct {
	client    *http.Client
	endpoint  string
	auth      string
	platform  string
	urlPrefix string
}

// A wrapper for http.NewRequestWithContext() that prepends tower endpoint to URL & sets authorization
// headers and then makes the actual http request.
func (c *AwxClient) GenericAPIRequest(ctx context.Context, method, url string, requestBody any, successCodes []int, aap25_api_endpoint_hint string) (responseBody []byte, statusCode int, errorMessage error) {

	url = c.buildAPIUrl(url, aap25_api_endpoint_hint)

	var body io.Reader

	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errorMessage = fmt.Errorf("unable to marshal requestBody into json: %s", err.Error())
			return
		}

		body = strings.NewReader(string(jsonData))
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		errorMessage = fmt.Errorf("error generating http request: %v", err)
		return
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", c.auth)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		errorMessage = fmt.Errorf("error doing http request: %v", err)
		return
	}

	var success bool
	for _, successCode := range successCodes {
		if httpResp.StatusCode == successCode {
			success = true
		}
	}

	responseBody, err = io.ReadAll(httpResp.Body)
	statusCode = httpResp.StatusCode

	if err != nil {
		errorMessage = fmt.Errorf("unable to read the http response data body. body: %v", responseBody)
		return
	}
	defer httpResp.Body.Close()

	if !success {
		errorMessage = fmt.Errorf("expected %v http response code for API call, got %d with message %s", successCodes, statusCode, responseBody)
		return
	}

	return
}

func (c *AwxClient) CreateUpdateAPIRequest(ctx context.Context, method, url string, requestBody any, successCodes []int, aap25_api_endpoint_hint string) (returnedData map[string]any, statusCode int, errorMessage error) {

	url = c.buildAPIUrl(url, aap25_api_endpoint_hint)

	var body io.Reader

	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errorMessage = fmt.Errorf("unable to marshal requestBody into json: %s", err.Error())
			return
		}

		body = strings.NewReader(string(jsonData))
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		errorMessage = fmt.Errorf("error generating http request: %v", err)
		return
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", c.auth)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		errorMessage = fmt.Errorf("error doing http request: %v", err)
		return
	}

	var success bool
	for _, successCode := range successCodes {
		if httpResp.StatusCode == successCode {
			success = true
		}
	}

	if !success {
		body, err := io.ReadAll(httpResp.Body)
		defer httpResp.Body.Close()
		if err != nil {
			errorMessage = errors.New("unable to read http request response body to retrieve error message")
			return
		}
		errorMessage = fmt.Errorf("expected %v http response code for API call, got %d with message %s", successCodes, httpResp.StatusCode, body)
		return
	}

	statusCode = httpResp.StatusCode
	httpRespBodyData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		errorMessage = errors.New("unable to read http request response body to retrieve id")
		return
	}
	err = json.Unmarshal(httpRespBodyData, &returnedData)
	if err != nil {
		errorMessage = errors.New("unable to unmarshal http request response body to retrieve returnedData")
		return
	}
	return
}

// In AAP, most api endpoint live in /controller/. But, sometimes they specifyc gateway endpoint instead.
func (c *AwxClient) buildAPIUrl(resourceUrl, aap25_api_endpoint_hint string) (url string) {

	if aap25_api_endpoint_hint == "gateway" && c.platform == "aap2.5" {
		url = c.endpoint + "/api/gateway/v1/" + resourceUrl
	} else {
		url = c.endpoint + c.urlPrefix + resourceUrl
	}

	return
}
