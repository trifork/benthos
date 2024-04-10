package amqp09

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/benthosdev/benthos/v4/public/service"
)

// OAuth2Config holds the configuration parameters for an OAuth2 exchange.
type OAuth2Config struct {
	Enabled      bool
	ClientId     string
	ClientSecret string
	TokenURL     string
	Scopes       []string
}

// NewOAuth2Config returns a new OAuth2Config with default values.
func newOAuth2Config() *OAuth2Config {
	return &OAuth2Config{
		Enabled:      false,
		ClientId:     "",
		ClientSecret: "",
		TokenURL:     "",
		Scopes:       []string{},
	}
}

func oauth2FromParsed(conf *service.ParsedConfig) (res *OAuth2Config, err error) {
	res = newOAuth2Config()
	if !conf.Contains(aFieldOAuth2) {
		return
	}
	conf = conf.Namespace(aFieldOAuth2)
	if res.Enabled, err = conf.FieldBool(ao2FieldEnabled); err != nil {
		return
	}
	if res.ClientId, err = conf.FieldString(ao2FieldClientId); err != nil {
		return
	}
	if res.ClientSecret, err = conf.FieldString(ao2FieldClientSecret); err != nil {
		return
	}
	if res.TokenURL, err = conf.FieldString(ao2FieldTokenURL); err != nil {
		return
	}
	if res.Scopes, err = conf.FieldStringList(ao2FieldScopes); err != nil {
		return
	}

	return
}

func acquireToken(ctx context.Context, c *OAuth2Config) (string, error) {

	authHeaderValue := base64.StdEncoding.EncodeToString([]byte(c.ClientId + ":" + c.ClientSecret))

	queryParams := url.Values{}
	queryParams.Set("grant_type", "client_credentials")
	queryParams.Set("scope", strings.Join(c.Scopes, " "))

	req, err := http.NewRequestWithContext(ctx, "POST", c.TokenURL, strings.NewReader(queryParams.Encode()))
	if err != nil {
		return "", err
	}

	req.URL.RawQuery = queryParams.Encode()

	req.Header.Set("Authorization", "Basic "+authHeaderValue)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := resp.Body.Close(); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status code %d", resp.StatusCode)
	}

	var tokenResponse map[string]interface{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse token response: %s", err)
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in token response")
	}

	return accessToken, nil
}
