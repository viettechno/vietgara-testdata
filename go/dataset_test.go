package testdata_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/viettechno/vietgara-testdata/go"
)

func TestLoadMergesFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeDataFile(t, dir, "environments.yaml", `version: 1
emailDomain: vietgara.test
environments:
  stubbed:
    password: local-password
    adminPassword: local-admin-password
`)
	writeDataFile(t, dir, "users.yaml", `users:
  - key: owner1
    name: QA Owner
    verified: true
`)
	writeDataFile(t, dir, "catalogue.yaml", `catalogue:
  requires: phase-6
  garage: gara1-1
  units:
    - key: litre
      name: Litre
      abbreviation: L
  items:
    - sku: OIL-5W30
      name: Engine oil
      baseUnit: litre
      packaging:
        - unit: can
          factor: 4
          salePrice: 380000
      salePrice: 120000
`)

	got, err := testdata.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Version != 1 || got.EmailDomain != "vietgara.test" {
		t.Fatalf("Load() dataset identity = (%d, %q), want (1, %q)", got.Version, got.EmailDomain, "vietgara.test")
	}
	if got.Environments["stubbed"].Password != "local-password" {
		t.Errorf("stubbed password = %q, want local-password", got.Environments["stubbed"].Password)
	}
	if len(got.Users) != 1 || got.Users[0].Key != "owner1" || got.Users[0].Verified == nil || !*got.Users[0].Verified {
		t.Errorf("users = %#v, want one verified owner1", got.Users)
	}
	if got.Catalogue == nil || len(got.Catalogue.Items) != 1 ||
		len(got.Catalogue.Items[0].Packaging) != 1 ||
		got.Catalogue.Items[0].Packaging[0].Factor != 4 ||
		got.Catalogue.Items[0].SalePrice != 120000 ||
		got.Catalogue.Items[0].Packaging[0].SalePrice == nil || *got.Catalogue.Items[0].Packaging[0].SalePrice != 380000 {
		t.Errorf("catalogue = %#v, want one item with packaging factor 4 and sale prices", got.Catalogue)
	}
}

func TestLoadRejectsInvalidFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantText string
	}{
		{
			name:     "no yaml files",
			files:    map[string]string{"notes.txt": "not a dataset"},
			wantText: "contains no .yaml files",
		},
		{
			name:     "empty yaml file",
			files:    map[string]string{"empty.yaml": ""},
			wantText: "expected a YAML mapping",
		},
		{
			name:     "non-mapping document",
			files:    map[string]string{"data.yaml": "- item"},
			wantText: "expected a YAML mapping",
		},
		{
			name:     "malformed yaml",
			files:    map[string]string{"data.yaml": "users: ["},
			wantText: "parse dataset file",
		},
		{
			name: "duplicate section across files",
			files: map[string]string{
				"a.yaml": "users: []\n",
				"b.yaml": "users: []\n",
			},
			wantText: `dataset section "users" is defined in both`,
		},
		{
			name: "unknown field",
			files: map[string]string{
				"users.yaml": "users:\n  - key: owner1\n    name: Owner\n    emali: typo@example.test\n",
			},
			wantText: "field emali not found",
		},
		{
			name:     "multiple documents",
			files:    map[string]string{"data.yaml": "users: []\n---\nusers: []\n"},
			wantText: "multiple YAML documents are not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for name, content := range tt.files {
				writeDataFile(t, dir, name, content)
			}

			_, err := testdata.Load(dir)
			if err == nil || !strings.Contains(err.Error(), tt.wantText) {
				t.Fatalf("Load() error = %v, want error containing %q", err, tt.wantText)
			}
		})
	}
}

func TestLoadRepositoryDataset(t *testing.T) {
	t.Parallel()

	got, err := testdata.Load("../data")
	if err != nil {
		t.Fatalf("Load(../data) error = %v", err)
	}
	if got.Version != 1 || len(got.Users) == 0 || len(got.Garages) == 0 {
		t.Fatalf("Load(../data) returned incomplete dataset: version=%d users=%d garages=%d",
			got.Version, len(got.Users), len(got.Garages))
	}
}

func ExampleLoad() {
	dataset, err := testdata.Load("../data")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("version=%d users=%d garages=%d\n",
		dataset.Version, len(dataset.Users), len(dataset.Garages))
	// Output: version=1 users=13 garages=3
}

func writeDataFile(t *testing.T, dir, name, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestLoadReadsLoadProfile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeDataFile(t, dir, "environments.yaml", `version: 1
emailDomain: vietgara.test
environments: {}
`)
	writeDataFile(t, dir, "load.yaml", `load:
  requires: phase-7
  profiles:
    - { key: garages50, garages: 50, ownerPlan: pro, customersPerGarage: 30, vehiclesPerCustomer: 2, openingStock: 5000, unitCost: 60000 }
  familyNames: [Nguyen]
  givenNames: [An]
  vehicles:
    - { make: Toyota, model: Vios, year: 2020 }
  parts:
    - { sku: LD-OIL, name: Engine oil, salePrice: 120000 }
  jobs:
    - { name: Oil change, price: 150000 }
`)

	got, err := testdata.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Load == nil || len(got.Load.Profiles) != 1 {
		t.Fatalf("Load = %#v, want one profile", got.Load)
	}
	profile := got.Load.Profiles[0]
	if profile.Garages != 50 || profile.OwnerPlan != "pro" || profile.OpeningStock != 5000 || profile.UnitCost != 60000 {
		t.Errorf("profile = %#v", profile)
	}
	if len(got.Load.Parts) != 1 || got.Load.Parts[0].SalePrice != 120000 || len(got.Load.Jobs) != 1 || got.Load.Jobs[0].Price != 150000 {
		t.Errorf("pools = %#v", got.Load)
	}
}
