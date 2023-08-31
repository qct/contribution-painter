package graphql

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rewriting-history/internal/pkg/helper"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type GraphClient struct {
	Url     string
	GhToken string
	Client  *http.Client
}

func NewClient(url, ghToken string, timeout time.Duration) *GraphClient {
	return &GraphClient{
		Url:     url,
		GhToken: ghToken,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *GraphClient) GraphQLRequest(query string, respContainer any) error {
	// Create the GraphQL request payload
	reqJSON, err := json.Marshal(graphqlRequest{
		Query: query,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP POST request to the GraphQL API
	req, err := http.NewRequest("POST", c.Url+helper.GitHubGraphQLPath, strings.NewReader(string(reqJSON)))
	if err != nil {
		log.Fatal("Failed to create HTTP request:", err)
	}
	// Set the necessary headers, including the access token
	req.Header.Set("Authorization", "Bearer "+c.GhToken)
	req.Header.Set("Content-Type", helper.ContentTypeJSON)
	req.Header.Set("Accept", helper.ContentTypeJSON)

	// Send the HTTP request
	resp, err := c.Client.Do(req)
	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("failed to get response: status code %d, status %s", resp.StatusCode, resp.Status)
	}
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	// Parse the GraphQL response
	err = json.NewDecoder(resp.Body).Decode(&respContainer)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
