package supabase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PabloViniegra/parking-app/services/core/internal/identity/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/ids"
)

type Authenticator struct {
	url, key string
	client   *http.Client
}

func NewAuthenticator(url, key string, client *http.Client) *Authenticator {
	return &Authenticator{url: url, key: key, client: client}
}

func (a *Authenticator) Verify(ctx context.Context, token string) (string, error) {
	if a.url == "" || a.key == "" {
		return "", fmt.Errorf("auth is not configured")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(a.url, "/")+"/auth/v1/user", nil)
	if err != nil {
		return "", fmt.Errorf("create Auth request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("apikey", a.key)
	response, err := a.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("verify access token: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return "", ports.ErrInvalidToken
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth service returned HTTP %d", response.StatusCode)
	}
	var user struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(&user); err != nil {
		return "", fmt.Errorf("decode Auth response: %w", err)
	}
	if !ids.Valid(user.ID) || user.Role != "authenticated" {
		return "", ports.ErrInvalidToken
	}
	return strings.ToLower(user.ID), nil
}
