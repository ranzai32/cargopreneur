package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var uploadPath = regexp.MustCompile(`^/uploads/[0-9a-f]{32}\.(jpg|png|webp)$`)

type CreateSupplierBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Region      string `json:"region"`
}

type NewSupplier struct {
	Name        string
	Description string
	Region      string
}

type CreateProductBody struct {
	SupplierID  string          `json:"supplierId"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Tags        json.RawMessage `json:"tags"`
	Unit        string          `json:"unit"`
	RetailPrice json.RawMessage `json:"retailPrice"`
	PriceTiers  json.RawMessage `json:"priceTiers"`
	ImageURL    string          `json:"imageUrl"`
}

type NewProduct struct {
	SupplierID  uuid.UUID
	Name        string
	Description string
	Category    string
	Tags        []string
	Unit        string
	RetailPrice float64
	PriceTiers  []PriceTier
	ImageURL    string
}

func ParseCreateSupplier(body CreateSupplierBody) (NewSupplier, map[string]string) {
	errs := map[string]string{}
	name := strings.TrimSpace(body.Name)
	region := strings.TrimSpace(body.Region)
	if name == "" {
		errs["name"] = "Name is required"
	}
	if region == "" {
		errs["region"] = "Region is required"
	}
	if len(errs) > 0 {
		return NewSupplier{}, errs
	}
	return NewSupplier{
		Name:        name,
		Description: strings.TrimSpace(body.Description),
		Region:      region,
	}, nil
}

func ParseCreateProduct(body CreateProductBody) (NewProduct, map[string]string) {
	errs := map[string]string{}
	var input NewProduct
	supplierID := strings.TrimSpace(body.SupplierID)
	if supplierID == "" {
		errs["supplierId"] = "Choose a supplier"
	} else if id, err := uuid.Parse(supplierID); err != nil {
		errs["supplierId"] = "Must be a valid id"
	} else {
		input.SupplierID = id
	}
	input.Name = strings.TrimSpace(body.Name)
	if input.Name == "" {
		errs["name"] = "Name is required"
	}
	input.Description = strings.TrimSpace(body.Description)
	input.Category = strings.TrimSpace(body.Category)
	if input.Category == "" {
		errs["category"] = "Category is required"
	}
	input.Unit = strings.TrimSpace(body.Unit)
	if input.Unit == "" {
		errs["unit"] = "Unit is required"
	}
	tags, tagMessage := parseTags(body.Tags)
	if tagMessage != "" {
		errs["tags"] = tagMessage
	} else {
		input.Tags = tags
	}
	price, ok := parseNonNegativeNumber(body.RetailPrice)
	if !ok {
		errs["retailPrice"] = "Enter a retail price of 0 or more"
	} else {
		input.RetailPrice = price
	}
	tiers, tierErrs := parsePriceTiers(body.PriceTiers)
	for key, message := range tierErrs {
		errs[key] = message
	}
	image := strings.TrimSpace(body.ImageURL)
	if image != "" && !uploadPath.MatchString(image) {
		errs["imageUrl"] = "Upload a JPEG, PNG, or WebP photo"
	} else {
		input.ImageURL = image
	}
	if len(errs) > 0 {
		return input, errs
	}
	input.PriceTiers = tiers
	return input, nil
}

func parseTags(raw json.RawMessage) ([]string, string) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return []string{}, ""
	}
	var tags []string
	if err := json.Unmarshal(trimmed, &tags); err != nil {
		return nil, "Enter tags as a list"
	}
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, "Enter a word for each tag"
		}
		cleaned = append(cleaned, tag)
	}
	return cleaned, ""
}

func parseNonNegativeNumber(raw json.RawMessage) (float64, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return 0, false
	}
	if trimmed[0] == '"' || trimmed[0] == '[' || trimmed[0] == '{' {
		return 0, false
	}
	var value float64
	if err := json.Unmarshal(trimmed, &value); err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, false
	}
	return value, true
}

func parsePriceTiers(raw json.RawMessage) ([]PriceTier, map[string]string) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte("[]")) {
		return nil, map[string]string{"priceTiers": "Add at least one price tier"}
	}
	var rows []struct {
		MinQty    json.RawMessage `json:"minQty"`
		UnitPrice json.RawMessage `json:"unitPrice"`
	}
	if err := json.Unmarshal(trimmed, &rows); err != nil {
		return nil, map[string]string{"priceTiers": "Add at least one price tier"}
	}
	if len(rows) == 0 {
		return nil, map[string]string{"priceTiers": "Add at least one price tier"}
	}
	errs := map[string]string{}
	tiers := make([]PriceTier, 0, len(rows))
	seen := map[int]bool{}
	duplicate := false
	for i, row := range rows {
		var tier PriceTier
		minQty, ok := parsePositiveInt(row.MinQty)
		if !ok {
			errs[fmt.Sprintf("priceTiers.%d.minQty", i)] = "Enter a minimum quantity of at least 1"
		} else {
			tier.MinQty = minQty
			if seen[minQty] {
				duplicate = true
			}
			seen[minQty] = true
		}
		unitPrice, priceOK := parseNonNegativeNumber(row.UnitPrice)
		if !priceOK {
			errs[fmt.Sprintf("priceTiers.%d.unitPrice", i)] = "Enter a unit price of 0 or more"
		} else {
			tier.UnitPrice = unitPrice
		}
		tiers = append(tiers, tier)
	}
	if duplicate {
		errs["priceTiers"] = "Each minimum quantity can only be used once"
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return tiers, nil
}
