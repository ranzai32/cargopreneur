package domain

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	SortNewest    = "newest"
	SortName      = "name"
	SortPriceAsc  = "price_asc"
	SortPriceDesc = "price_desc"

	DefaultPage     = 1
	DefaultPageSize = 12
	MaxPageSize     = 50
)

type PriceTier struct {
	MinQty    int
	UnitPrice float64
}

type SupplierBrief struct {
	ID         uuid.UUID
	Name       string
	Region     string
	IsVerified bool
}

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	Tags        []string
	Unit        string
	RetailPrice float64
	Supplier    SupplierBrief
	PriceTiers  []PriceTier
	CreatedAt   time.Time
}

type Supplier struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Region       string
	IsVerified   bool
	ProductCount int
}

type Category struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type PriceTierJSON struct {
	MinQty    int     `json:"minQty"`
	UnitPrice float64 `json:"unitPrice"`
}

type SupplierBriefJSON struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	IsVerified bool      `json:"isVerified"`
}

type ProductJSON struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Unit        string            `json:"unit"`
	RetailPrice float64           `json:"retailPrice"`
	Supplier    SupplierBriefJSON `json:"supplier"`
	PriceTiers  []PriceTierJSON   `json:"priceTiers"`
	CreatedAt   time.Time         `json:"createdAt"`
}

type SupplierJSON struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Region       string    `json:"region"`
	IsVerified   bool      `json:"isVerified"`
	ProductCount int       `json:"productCount"`
}

type ProductQuery struct {
	Q                 string
	Category          string
	SupplierID        *uuid.UUID
	Region            string
	Sort              string
	Page              int
	PageSize          int
	IncludeUnverified bool
}

func (p Product) JSON() ProductJSON {
	tags := p.Tags
	if tags == nil {
		tags = []string{}
	}
	tiers := append([]PriceTier(nil), p.PriceTiers...)
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].MinQty < tiers[j].MinQty
	})
	if tiers == nil {
		tiers = []PriceTier{}
	}
	encodedTiers := make([]PriceTierJSON, len(tiers))
	for i, tier := range tiers {
		encodedTiers[i] = PriceTierJSON{MinQty: tier.MinQty, UnitPrice: tier.UnitPrice}
	}
	return ProductJSON{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Tags:        tags,
		Unit:        p.Unit,
		RetailPrice: p.RetailPrice,
		Supplier: SupplierBriefJSON{
			ID:         p.Supplier.ID,
			Name:       p.Supplier.Name,
			Region:     p.Supplier.Region,
			IsVerified: p.Supplier.IsVerified,
		},
		PriceTiers: encodedTiers,
		CreatedAt:  p.CreatedAt,
	}
}

func (s Supplier) JSON() SupplierJSON {
	return SupplierJSON{
		ID:           s.ID,
		Name:         s.Name,
		Description:  s.Description,
		Region:       s.Region,
		IsVerified:   s.IsVerified,
		ProductCount: s.ProductCount,
	}
}

func ParseProductQuery(values url.Values) (ProductQuery, map[string]string) {
	errs := map[string]string{}
	query := ProductQuery{Sort: SortNewest, Page: DefaultPage, PageSize: DefaultPageSize}
	query.Q = strings.TrimSpace(values.Get("q"))
	query.Category = strings.TrimSpace(values.Get("category"))
	query.Region = strings.TrimSpace(values.Get("region"))

	if raw := strings.TrimSpace(values.Get("supplierId")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			errs["supplierId"] = "Must be a valid id"
		} else {
			query.SupplierID = &id
		}
	}

	if raw := strings.TrimSpace(values.Get("sort")); raw != "" {
		switch raw {
		case SortNewest, SortName, SortPriceAsc, SortPriceDesc:
			query.Sort = raw
		default:
			errs["sort"] = "Use newest, name, price_asc, or price_desc"
		}
	}

	if page, message := parseQueryInt(values.Get("page"), 1, 0); message != "" {
		errs["page"] = message
	} else {
		query.Page = page
	}
	if pageSize, message := parseQueryInt(values.Get("pageSize"), DefaultPageSize, MaxPageSize); message != "" {
		errs["pageSize"] = message
	} else {
		query.PageSize = pageSize
	}
	return query, errs
}

func ParseIncludeUnverified(raw string) (bool, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, ""
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, "Use true or false"
	}
	return value, ""
}

func parseQueryInt(raw string, fallback, max int) (int, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, ""
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, "Must be a whole number of 1 or more"
	}
	if max > 0 && value > max {
		return 0, fmt.Sprintf("Must be at most %d", max)
	}
	return value, ""
}

func LowestUnitPrice(tiers []PriceTier) (float64, bool) {
	if len(tiers) == 0 {
		return 0, false
	}
	lowest := tiers[0].UnitPrice
	for _, tier := range tiers[1:] {
		if tier.UnitPrice < lowest {
			lowest = tier.UnitPrice
		}
	}
	return lowest, true
}
