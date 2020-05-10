package runner

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/go-github/v31/github"
)

const (
	urlPrefix = "https://github.com/"
)

// TokenProvider is simple github.Client wrapper responsible for creating registration and remove tokens for runners
type TokenProvider struct {
	baseURL string
	client  *github.Client
}

// NewTokenProvider returns new TokenProvider instance
func NewTokenProvider(url string, client *github.Client) (*TokenProvider, error) {
	if len(url) == 0 {
		return nil, errors.Wrap(ErrMissingParameter, "url is missing")
	}

	if client == nil {
		return nil, errors.Wrap(ErrMissingParameter, "client is missing")
	}

	provider := &TokenProvider{
		client: client,
	}

	owner, repo, err := splitURL(url)
	if err != nil {
		return nil, err
	}

	if len(repo) == 0 {
		provider.baseURL = fmt.Sprintf("/orgs/%s", owner)
	} else {
		provider.baseURL = fmt.Sprintf("/repos/%s/%s", owner, repo)
	}

	return provider, nil
}

//CreateRegistrationToken returns new runner registration token
func (tp *TokenProvider) CreateRegistrationToken() (*github.RegistrationToken, error) {
	url := fmt.Sprintf("%s/actions/runners/registration-token", tp.baseURL)

	var tok github.RegistrationToken

	err := createToken(tp.client, url, &tok)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create registration token for: %s", url)
	}

	return &tok, nil
}

//CreateRemoveToken returns new runner remove token
func (tp *TokenProvider) CreateRemoveToken() (*github.RemoveToken, error) {
	url := fmt.Sprintf("%s/actions/runners/remove-token", tp.baseURL)

	var tok github.RemoveToken

	err := createToken(tp.client, url, &tok)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create remove token for: %s", url)
	}

	return &tok, nil
}

func createToken(client *github.Client, url string, tok interface{}) error {
	// create http request
	req, err := client.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	// execute the request
	resp, err := client.Do(context.Background(), req, tok)
	if err != nil {
		return err
	}

	// check if token created successfully
	if resp.StatusCode != http.StatusCreated {
		return errors.Wrapf(ErrGHRequestFailed, "unexpected status: %s", resp.Status)
	}

	return nil
}

func splitURL(url string) (string, string, error) {
	// validate url is a github url
	if !strings.HasPrefix(url, urlPrefix) {
		return "", "", errors.Wrapf(ErrInvalidParameter, "invalid url: '%s'", url)
	}

	// get url path and split chunks
	path := strings.TrimPrefix(url, urlPrefix)
	chunk := strings.Split(path, "/")

	//nolint
	switch len(chunk) {
	case 1: // organization url = https://github.com/foo-org
		return chunk[0], "", nil
	case 2: // owner and repo url https://github.com/foo/bar-repo
		return chunk[0], chunk[1], nil
	default: // invalid url path
		return "", "", fmt.Errorf("invalid url path: '%s'", path)
	}
}
