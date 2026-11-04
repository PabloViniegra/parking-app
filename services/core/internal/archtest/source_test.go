package archtest

import "testing"

func TestCheckSource(t *testing.T) {
	const domain = "github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	const adapter = "github.com/PabloViniegra/parking-app/services/core/internal/availability/adapters/postgres"

	tests := []struct {
		name      string
		pkg       string
		src       string
		wantCount int
	}{
		{
			name: "domain using time types is allowed",
			pkg:  domain,
			src:  "package domain\nimport \"time\"\nfunc ttl(at time.Time) time.Time { return at.Add(180 * time.Second) }",
		},
		{
			name:      "domain calling time.Now is rejected",
			pkg:       domain,
			src:       "package domain\nimport \"time\"\nfunc now() time.Time { return time.Now() }",
			wantCount: 1,
		},
		{
			name:      "domain calling time.Now through an alias is rejected",
			pkg:       domain,
			src:       "package domain\nimport t \"time\"\nvar f = t.Now",
			wantCount: 1,
		},
		{
			name:      "domain importing another module's domain is rejected",
			pkg:       domain,
			src:       "package domain\nimport _ \"github.com/PabloViniegra/parking-app/services/core/internal/matching/domain\"",
			wantCount: 1,
		},
		{
			name: "adapter calling time.Now is allowed",
			pkg:  adapter,
			src:  "package postgres\nimport \"time\"\nvar f = time.Now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckSource(tt.pkg, "file.go", []byte(tt.src))
			if err != nil {
				t.Fatalf("CheckSource() unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("CheckSource() = %v, want %d violations", got, tt.wantCount)
			}
		})
	}
}
