package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type SourcesService interface {
	GetApplicationTypeIDs(ctx context.Context) (string, error)
	GetApplicationByID(id, ebsAccount string) (*Application, error)
	GetAuthenticationByID(id, ebsAccount string) (*Authentication, error)
}

type SourceServiceImpl struct {
	client  *http.Client
	baseUrl string
	psk     string
}

func New(client *http.Client, baseUrl, psk string) SourceServiceImpl {
	return SourceServiceImpl{client: client, baseUrl: baseUrl, psk: psk}
}

func (s *SourceServiceImpl) GetApplicationTypeIDs(ctx context.Context) (string, error) {
	appName := "/insights/platform/cloud-meter"
	url := fmt.Sprintf("%s%s", s.baseUrl, "/application_types")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	var appIds ApplicationTypeIDs
	_ = json.NewDecoder(resp.Body).Decode(&appIds)
	for _, appType := range appIds.Data {
		if appType.Name == appName {
			return appType.ID, nil
		}
	}
	return "", errors.New("not found")
}

func (s *SourceServiceImpl) GetApplicationByID(id, ebsAccount string) (*Application, error) {
	url := fmt.Sprintf("%s%s%s", s.baseUrl, "/applications/", id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-rh-sources-account-number", ebsAccount)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	var app Application
	_ = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &app, nil
}

func (s *SourceServiceImpl) GetAuthenticationByID(id, ebsAccount string) (*Authentication, error) {
	url := fmt.Sprintf("%s%s%s", s.baseUrl, "/authentications/", id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-rh-sources-account-number", ebsAccount)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	var auth Authentication
	_ = json.NewDecoder(resp.Body).Decode(&auth)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &auth, nil
}
