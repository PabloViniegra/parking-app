package bootstrap

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDatabaseConfig(databaseURL, caPEM string) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database connection: %w", err)
	}
	if caPEM == "" {
		return config, nil
	}
	tls := config.ConnConfig.TLSConfig
	if tls == nil || tls.InsecureSkipVerify || tls.ServerName == "" {
		return nil, errors.New("configured database CA requires sslmode=verify-full")
	}
	ca := x509.NewCertPool()
	if !ca.AppendCertsFromPEM([]byte(caPEM)) {
		return nil, errors.New("invalid database CA certificate")
	}
	tls.RootCAs = ca
	return config, nil
}
