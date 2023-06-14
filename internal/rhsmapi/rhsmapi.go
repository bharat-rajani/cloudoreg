package rhsmapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const accountAuthHeader = "X-RhsmApi-AccountID"
const callContextHeader = "X-RhApiPlatform-CallContext"

var (
	autoRegRoles = []string{
		"api_access",
		"account_auth",
		"cloud_access_auto_registration",
	}

	verified = true
)

type AccountDetails struct {
	Provider          string `json:"-"`
	ProviderAccountID string `json:"id"`
	Nickname          string `json:"nickname"`
	Verified          *bool  `json:"verified"`
	SourceID          string `json:"sourceId"`
	AccountID         string `json:"-"`
}

type Broker struct {
	baseURL string
	client  http.Client
}

func (b *Broker) CreateAccountWithAutoReg(acc AccountDetails) error {
	endpoint := fmt.Sprintf("/cloud_access_providers/%s/accounts", acc.Provider)
	reqBody := []AccountDetails{
		{
			ProviderAccountID: acc.ProviderAccountID,
			Verified:          &verified,
			Nickname:          acc.Nickname,
			SourceID:          acc.SourceID,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, b.baseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add(accountAuthHeader, acc.AccountID)
	// TODO: Remove once advanced to active development
	req.Header.Add(callContextHeader, mockSidecarHeader(autoRegRoles))

	res, err := b.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("create accounts failed with status code %d", res.StatusCode)
	}

	return nil
}

func NewBroker(url string) *Broker {
	return &Broker{
		baseURL: url,
		// TODO: Add support for TLS
		client: http.Client{},
	}
}

func mockSidecarHeader(clientRoles []string) string {
	roles := strings.Join(clientRoles, `", "`)
	return base64.StdEncoding.EncodeToString([]byte(`{
		"client":{"serviceAccountId":"pat",
		"roles":["` + roles + `"]}}`))
}
