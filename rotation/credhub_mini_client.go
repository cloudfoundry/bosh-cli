package rotation

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type credhubClient interface {
	ListCredentials(prefix string) (*credhubListResponse, error)
	GetCredential(name string) (*credhubCredResp, error)
	SetValueCredential(name, value string) error

	// GetCredentialValue looks for a value, returning default if not found (and nil error)
	// Other errors will be returned
	GetCredentialValue(name, defaultValue string) (string, error)
	DeleteCredential(name string) error
	GetCertificateID(name string) (string, error)
	MakeTransitionalCertificate(certificateID string) (*credVersionData, error)
	MakeThisOneTransitional(certificateID, otherOneToMakeTransitionalOrBlankForNone string) error
}

type credhubListResponse struct {
	Credentials []struct {
		Name string `json:"name"`
	} `json:"credentials"`
}

type credVersionData struct {
	VersionCreatedAt time.Time   `json:"version_created_at"`
	Type             string      `json:"type"`
	ID               string      `json:"id"`
	Transitional     bool        `json:"transitional"`
	Value            interface{} `json:"value"` // poor choice of API surface...
}

type credhubCredResp struct {
	Data []*credVersionData `json:"data"`
}

func newCredhubClient(credhubCerts, credhubBaseURL, credhubClient, credhubClientSecret string) (credhubClient, error) {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM([]byte(credhubCerts)) {
		return nil, errors.New("CREDHUB_CA_CERT must be set and include PEMs")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: cp,
			},
		},
	}

	uaaURL, err := getUAAURL(client, credhubBaseURL)
	if err != nil {
		return nil, err
	}

	token, err := getUAAToken(client, uaaURL, credhubClient, credhubClientSecret)
	if err != nil {
		return nil, err
	}

	return &miniCredhubClient{
		AccessToken: token,
		Client:      client,
		BaseURL:     credhubBaseURL,
	}, nil
}

type miniCredhubClient struct {
	AccessToken string
	Client      *http.Client
	BaseURL     string
}

func (chc *miniCredhubClient) MakeThisOneTransitional(certificateID, otherOneToMakeTransitionalOrBlankForNone string) error {
	var reqData []byte
	if otherOneToMakeTransitionalOrBlankForNone == "" {
		reqData = []byte(`{"version":null}`)
	} else {
		var err error
		reqData, err = json.Marshal(&(struct {
			S string `json:"version"`
		}{S: otherOneToMakeTransitionalOrBlankForNone}))
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/certificates/%s/update_transitional_version", chc.BaseURL, certificateID), bytes.NewReader(reqData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad status code")
	}

	return nil
}

func (chc *miniCredhubClient) MakeTransitionalCertificate(certificateID string) (*credVersionData, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/certificates/%s/regenerate", chc.BaseURL, certificateID), bytes.NewReader([]byte(`{"set_as_transitional":true}`)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status code")
	}

	var x credVersionData
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (chc *miniCredhubClient) GetCertificateID(name string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/certificates?%s", chc.BaseURL, (&url.Values{
		"name": []string{name},
	}).Encode()), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status code")
	}

	var x struct {
		Certificates []struct {
			ID string `json:"id"`
		} `json:"certificates"`
	}
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return "", err
	}

	if len(x.Certificates) != 1 {
		return "", errors.New("wrong num certs")
	}

	if x.Certificates[0].ID == "" {
		return "", errors.New("no cert ID")
	}

	return x.Certificates[0].ID, nil
}

func (chc *miniCredhubClient) DeleteCredential(name string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/data?%s", chc.BaseURL, (&url.Values{
		"name": []string{name},
	}).Encode()), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("bad status code")
	}

	return nil
}

func (chc *miniCredhubClient) SetValueCredential(name, value string) error {
	rd, err := json.Marshal(struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Name:  name,
		Type:  "value",
		Value: value,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/data", chc.BaseURL), bytes.NewReader(rd))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad status code")
	}

	return nil
}

func (chc *miniCredhubClient) GetCredentialValue(name, defaultValue string) (string, error) {
	cr, err := chc.GetCredential(name)
	switch err {
	case nil:
		if len(cr.Data) != 1 {
			return "", errors.New("unexpected number of results")
		}
		rv, ok := cr.Data[0].Value.(string)
		if !ok {
			return "", errors.New("unable to coerce data to string")
		}
		return rv, nil

	case errNotFound:
		return defaultValue, nil

	default:
		return "", err
	}
}

var (
	errNotFound = errors.New("not found")
)

func (chc *miniCredhubClient) GetCredential(name string) (*credhubCredResp, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/data?%s", chc.BaseURL, (&url.Values{
		"name":    []string{name},
		"current": []string{"true"},
	}).Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errNotFound
		}
		return nil, errors.New("bad status code")
	}

	var x credhubCredResp
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func (chc *miniCredhubClient) ListCredentials(path string) (*credhubListResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/data?%s", chc.BaseURL, (&url.Values{
		"path": []string{path},
	}).Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", chc.AccessToken))
	resp, err := chc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad status code")
	}

	var x credhubListResponse
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

func getUAAToken(client *http.Client, uaaURL, clientID, clientSecret string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/oauth/token", uaaURL), bytes.NewReader([]byte((&url.Values{
		"client_id":     []string{clientID},
		"client_secret": []string{clientSecret},
		"grant_type":    []string{"client_credentials"},
		"token_type":    []string{"jwt"},
	}).Encode())))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status code")
	}

	var x struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return "", err
	}

	if x.AccessToken == "" {
		return "", errors.New("no access token found in response")
	}
	return x.AccessToken, nil
}

func getUAAURL(client *http.Client, baseURL string) (string, error) {
	resp, err := client.Get(fmt.Sprintf("%s/info", baseURL))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad status code")
	}

	var x struct {
		AuthServer struct {
			URL string `json:"url"`
		} `json:"auth-server"`
	}
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return "", err
	}

	if x.AuthServer.URL == "" {
		return "", errors.New("no auth server URL found in response")
	}

	return x.AuthServer.URL, nil
}
