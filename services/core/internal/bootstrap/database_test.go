package bootstrap_test

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/bootstrap"
)

func TestDatabaseConfig_TrustsConfiguredCAWithoutDisablingVerification(t *testing.T) {
	ca, err := os.ReadFile("testdata/supabase-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	config, err := bootstrap.NewDatabaseConfig("postgresql://runtime@pooler.example.test/postgres?sslmode=verify-full", string(ca))
	if err != nil {
		t.Fatal(err)
	}
	tls := config.ConnConfig.TLSConfig
	if tls == nil || tls.InsecureSkipVerify || tls.ServerName != "pooler.example.test" || tls.RootCAs == nil {
		t.Fatalf("CA or hostname verification is missing")
	}
	block, _ := pem.Decode(ca)
	if block == nil {
		t.Fatal("invalid CA fixture")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := certificate.Verify(x509.VerifyOptions{Roots: tls.RootCAs, CurrentTime: time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("configured Supabase CA is not trusted: %v", err)
	}
}

func TestDatabaseConfig_RejectsInvalidOrUnverifiedCustomCA(t *testing.T) {
	ca, err := os.ReadFile("testdata/supabase-ca.crt")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ name, mode, certificate string }{
		{name: "invalid PEM", mode: "verify-full", certificate: "invalid"},
		{name: "unverified TLS", mode: "require", certificate: string(ca)},
		{name: "plaintext", mode: "disable", certificate: string(ca)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := bootstrap.NewDatabaseConfig("postgresql://runtime@pooler.example.test/postgres?sslmode="+tt.mode, tt.certificate); err == nil {
				t.Fatal("accepted invalid custom CA configuration")
			}
		})
	}
}
