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
func (c *AwxClient) GenericAPIRequest(ctx context.Context, method, url string, requestBody any, successCodes []int) (responseBody []byte, statusCode int, errorMessage error) {

	url = c.buildAPIUrl(ctx, url, method)

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

func (c *AwxClient) CreateUpdateAPIRequest(ctx context.Context, method, url string, requestBody any, successCodes []int) (returnedData map[string]any, statusCode int, errorMessage error) {

	url = c.buildAPIUrl(ctx, url, method)

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

// In AAP, most api endpoint live in /controller/. But, for the few exceptions, in list below, override to /gateway/v1/.
func (c *AwxClient) buildAPIUrl(ctx context.Context, resourceUrl, httpMethod string) (url string) {

	var contextKey contextKey = "dataSource"
	dataSource := ctx.Value(contextKey)

	aap_gateway_override_cud_list := []string{"organizations"}

	if c.platform != "awx" && c.platform != "aap2.4" && httpMethod != http.MethodGet {
		for _, v := range aap_gateway_override_cud_list {
			if strings.HasPrefix(resourceUrl, v) {
				url = c.endpoint + "/api/gateway/v1/" + resourceUrl
				return
			}
		}
	}

	aap_gateway_override_all_list := []string{"users"}

	if c.platform != "awx" && c.platform != "aap2.4" {
		for _, v := range aap_gateway_override_all_list {
			if strings.HasPrefix(resourceUrl, v) {
				url = c.endpoint + "/api/gateway/v1/" + resourceUrl
				return
			}
		}
	}

	// use controller for data source URLs, but not in resources
	aap_gateway_override_non_data_source := []string{"organizations/?name"}

	if c.platform != "awx" && c.platform != "aap2.4" && dataSource == nil {
		for _, v := range aap_gateway_override_non_data_source {
			if strings.HasPrefix(resourceUrl, v) {
				url = c.endpoint + "/api/gateway/v1/" + resourceUrl
				return
			}
		}
	}

	url = c.endpoint + c.urlPrefix + resourceUrl

	return
}
