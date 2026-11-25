package places

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TwoGIS struct {
	Key    string
	Client *http.Client
}

type suggestResponse struct {
	Result struct {
		Items []struct {
			FullName        string `json:"full_name"`
			AddressName     string `json:"address_name"`
			FullAddressName string `json:"full_address_name"`
		} `json:"items"`
	} `json:"result"`
}

func (t TwoGIS) Suggest(ctx context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if t.Key == "" || len([]rune(query)) < 3 {
		return []string{}, nil
	}
	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 4 * time.Second}
	}
	endpoint, err := url.Parse("https://catalog.api.2gis.com/3.0/suggests")
	if err != nil {
		return nil, err
	}
	values := endpoint.Query()
	values.Set("q", query)
	values.Set("key", t.Key)
	values.Set("suggest_type", "address")
	values.Set("locale", "ru_KZ")
	values.Set("page_size", "6")
	endpoint.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("2gis status %d", res.StatusCode)
	}
	var payload suggestResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return suggestionLabels(payload), nil
}

func suggestionLabels(payload suggestResponse) []string {
	seen := map[string]struct{}{}
	items := make([]string, 0, len(payload.Result.Items))
	for _, item := range payload.Result.Items {
		label := strings.TrimSpace(item.FullAddressName)
		if label == "" {
			label = strings.TrimSpace(item.FullName)
		}
		if label == "" {
			label = strings.TrimSpace(item.AddressName)
		}
		if label == "" {
			continue
		}
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		items = append(items, label)
	}
	return items
}
