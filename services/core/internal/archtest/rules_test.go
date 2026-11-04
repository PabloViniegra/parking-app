package archtest

import "testing"

func TestCheckImport(t *testing.T) {
	const m = "github.com/PabloViniegra/parking-app/services/core/internal/"

	tests := []struct {
		name    string
		from    string
		imp     string
		wantErr bool
	}{
		{name: "domain may import stdlib", from: m + "availability/domain", imp: "time"},
		{name: "domain may import shared events", from: m + "availability/domain", imp: m + "shared/events"},
		{name: "domain must not import shared clock", from: m + "availability/domain", imp: m + "shared/clock", wantErr: true},
		{name: "domain must not import its ports", from: m + "availability/domain", imp: m + "availability/ports", wantErr: true},
		{name: "domain must not import third-party", from: m + "availability/domain", imp: "github.com/uber/h3-go/v4", wantErr: true},
		{name: "domain must not import net/http", from: m + "availability/domain", imp: "net/http", wantErr: true},

		{name: "application may import own domain", from: m + "matching/application", imp: m + "matching/domain"},
		{name: "application may import own ports", from: m + "matching/application", imp: m + "matching/ports"},
		{name: "application may import shared", from: m + "matching/application", imp: m + "shared/clock"},
		{name: "application must not import adapters", from: m + "matching/application", imp: m + "matching/adapters/postgres", wantErr: true},
		{name: "application must not import database/sql", from: m + "matching/application", imp: "database/sql", wantErr: true},
		{name: "application must not import net/http", from: m + "matching/application", imp: "net/http", wantErr: true},
		{name: "application must not import third-party", from: m + "matching/application", imp: "github.com/jackc/pgx/v5", wantErr: true},

		{name: "ports may import own domain", from: m + "matching/ports", imp: m + "matching/domain"},
		{name: "ports must not import adapters", from: m + "matching/ports", imp: m + "matching/adapters/http", wantErr: true},

		{name: "adapters may import third-party", from: m + "zones/adapters/postgres", imp: "github.com/jackc/pgx/v5"},
		{name: "adapters may import own application", from: m + "zones/adapters/http", imp: m + "zones/application"},

		{name: "module may import another module's ports", from: m + "matching/application", imp: m + "availability/ports"},
		{name: "module must not import another module's domain", from: m + "matching/application", imp: m + "availability/domain", wantErr: true},
		{name: "module must not import another module's adapters", from: m + "matching/adapters/postgres", imp: m + "availability/adapters/postgres", wantErr: true},

		{name: "shared must not import business modules", from: m + "shared/outbox", imp: m + "availability/domain", wantErr: true},
		{name: "bootstrap may wire any module", from: m + "bootstrap", imp: m + "availability/adapters/http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckImport(tt.from, tt.imp)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckImport(%q, %q) error = %v, wantErr %v", tt.from, tt.imp, err, tt.wantErr)
			}
		})
	}
}
