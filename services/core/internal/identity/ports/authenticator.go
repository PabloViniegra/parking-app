package ports

import (
	"context"
	"errors"
)

var ErrInvalidToken = errors.New("invalid access token")

type Authenticator interface {
	Verify(context.Context, string) (string, error)
}
