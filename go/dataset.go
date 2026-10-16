// Package testdata loads the shared VietGara test dataset from its YAML files.
package testdata

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Dataset is the typed representation of the merged data/*.yaml files.
type Dataset struct {
	Version      int                    `yaml:"version"`
	EmailDomain  string                 `yaml:"emailDomain"`
	Environments map[string]Environment `yaml:"environments"`
	Plans        []Plan                 `yaml:"plans"`
	Users        []User                 `yaml:"users"`
	Garages      []Garage               `yaml:"garages"`
	Roles        []Role                 `yaml:"roles"`
	Groups       []Group                `yaml:"groups"`
	Staff        []Staff                `yaml:"staff"`
	Customers    []Customer             `yaml:"customers"`
	Suppliers    []Supplier             `yaml:"suppliers"`
	Catalogue    *Catalogue             `yaml:"catalogue"`
}

// Environment contains either local credentials or staging secret references.
type Environment struct {
	Password         string `yaml:"password"`
	AdminPassword    string `yaml:"adminPassword"`
	PasswordRef      string `yaml:"passwordRef"`
	AdminPasswordRef string `yaml:"adminPasswordRef"`
}

// Plan describes a test subscription plan.
type Plan struct {
	Code         string   `yaml:"code"`
	Name         string   `yaml:"name"`
	Price        int      `yaml:"price"`
	Currency     string   `yaml:"currency"`
	BillingCycle string   `yaml:"billingCycle"`
	Limits       Limits   `yaml:"limits"`
	Features     []string `yaml:"features"`
}

// Limits contains the usage limits for a plan.
type Limits struct {
	MaxGarages            int `yaml:"maxGarages"`
	MaxStaff              int `yaml:"maxStaff"`
	MaxTransactionsPerDay int `yaml:"maxTransactionsPerDay"`
}

// User describes a platform, owner, staff, or supplier-portal user.
type User struct {
	Key           string `yaml:"key"`
	Name          string `yaml:"name"`
	Phone         string `yaml:"phone"`
	Plan          string `yaml:"plan"`
	PlatformAdmin bool   `yaml:"platformAdmin"`
	PlatformRole  string `yaml:"platformRole"`
	SupplierFor   string `yaml:"supplierFor"`
	Verified      *bool  `yaml:"verified"`
	Requires      string `yaml:"requires"`
	Note          string `yaml:"note"`
}

// Garage describes a garage owned by a dataset user.
type Garage struct {
	Key     string `yaml:"key"`
	Owner   string `yaml:"owner"`
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Phone   string `yaml:"phone"`
	TaxCode string `yaml:"taxCode"`
}

// Role describes a custom role owned by a user.
type Role struct {
	Owner       string   `yaml:"owner"`
	Name        string   `yaml:"name"`
	Permissions []string `yaml:"permissions"`
}

// Group assigns a role to a garage.
type Group struct {
	Garage string `yaml:"garage"`
	Name   string `yaml:"name"`
	Role   string `yaml:"role"`
}

// Staff assigns a user to a garage and group.
type Staff struct {
	User   string `yaml:"user"`
	Garage string `yaml:"garage"`
	Group  string `yaml:"group"`
}

// Customer describes a garage customer and their vehicles.
type Customer struct {
	Key      string    `yaml:"key"`
	Garage   string    `yaml:"garage"`
	Name     string    `yaml:"name"`
	Phone    string    `yaml:"phone"`
	Email    string    `yaml:"email"`
	Consent  *bool     `yaml:"consent"`
	Vehicles []Vehicle `yaml:"vehicles"`
}

// Vehicle describes a vehicle belonging to a customer.
type Vehicle struct {
	Plate string `yaml:"plate"`
	Make  string `yaml:"make"`
	Model string `yaml:"model"`
	Year  int    `yaml:"year"`
}

// Supplier describes a supplier associated with a garage.
type Supplier struct {
	Key     string          `yaml:"key"`
	Garage  string          `yaml:"garage"`
	Name    string          `yaml:"name"`
	TaxCode string          `yaml:"taxCode"`
	Contact SupplierContact `yaml:"contact"`
}

// SupplierContact contains a supplier's test contact details.
type SupplierContact struct {
	Person string `yaml:"person"`
	Phone  string `yaml:"phone"`
	Email  string `yaml:"email"`
}

// Catalogue contains a garage's test units, categories, and parts.
type Catalogue struct {
	Requires   string     `yaml:"requires"`
	Garage     string     `yaml:"garage"`
	Units      []Unit     `yaml:"units"`
	Categories []Category `yaml:"categories"`
	Items      []Item     `yaml:"items"`
}

// Unit describes an inventory unit of measure.
type Unit struct {
	Key          string `yaml:"key"`
	Name         string `yaml:"name"`
	Abbreviation string `yaml:"abbreviation"`
}

// Category is a possibly nested part category.
type Category struct {
	Key      string     `yaml:"key"`
	Name     string     `yaml:"name"`
	Children []Category `yaml:"children"`
}

// Item describes a part in the test catalogue.
type Item struct {
	SKU              string      `yaml:"sku"`
	Name             string      `yaml:"name"`
	Category         string      `yaml:"category"`
	Brand            string      `yaml:"brand"`
	BaseUnit         string      `yaml:"baseUnit"`
	Packaging        []Packaging `yaml:"packaging"`
	ReorderThreshold int         `yaml:"reorderThreshold"`
	// SalePrice is VND per base unit (whole dong).
	SalePrice int `yaml:"salePrice"`
	DefaultSupplier  string      `yaml:"defaultSupplier"`
}

// Packaging describes an inventory unit and its conversion factor.
type Packaging struct {
	Unit   string  `yaml:"unit"`
	Factor float64 `yaml:"factor"`
	// SalePrice overrides base price times factor for one packaging unit; nil derives it.
	SalePrice *int `yaml:"salePrice"`
}

// Load reads and merges the *.yaml files in dataDir into a typed Dataset.
// Unknown fields and duplicate top-level sections are rejected to catch
// drift between the dataset and this package's types. Use the repository's
// make validate target for JSON Schema and cross-reference validation.
func Load(dataDir string) (*Dataset, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("read dataset directory %q: %w", dataDir, err)
	}

	merged := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	sectionFiles := make(map[string]string)
	fileCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		fileCount++

		path := filepath.Join(dataDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read dataset file %q: %w", path, err)
		}

		root, err := readDocument(path, content)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(root.Content); i += 2 {
			key := root.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return nil, fmt.Errorf("dataset file %q: top-level keys must be strings", path)
			}
			name := key.Value
			if previous, ok := sectionFiles[name]; ok {
				return nil, fmt.Errorf("dataset section %q is defined in both %q and %q", name, previous, path)
			}
			sectionFiles[name] = path
			merged.Content = append(merged.Content, key, root.Content[i+1])
		}
	}

	if fileCount == 0 {
		return nil, fmt.Errorf("dataset directory %q contains no .yaml files", dataDir)
	}

	document, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("merge dataset files in %q: %w", dataDir, err)
	}

	var dataset Dataset
	decoder := yaml.NewDecoder(bytes.NewReader(document))
	decoder.KnownFields(true)
	if err := decoder.Decode(&dataset); err != nil {
		return nil, fmt.Errorf("decode dataset in %q: %w", dataDir, err)
	}

	return &dataset, nil
}

func readDocument(path string, content []byte) (*yaml.Node, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("parse dataset file %q: expected a YAML mapping", path)
		}
		return nil, fmt.Errorf("parse dataset file %q: %w", path, err)
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("parse dataset file %q: expected a YAML mapping", path)
	}

	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("parse dataset file %q: %w", path, err)
		}
		return nil, fmt.Errorf("parse dataset file %q: multiple YAML documents are not supported", path)
	}

	root := document.Content[0]
	if len(root.Content)%2 != 0 {
		return nil, fmt.Errorf("parse dataset file %q: malformed YAML mapping", path)
	}
	return root, nil
}
