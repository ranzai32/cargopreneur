package places

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Suggester interface {
	Suggest(ctx context.Context, query string) ([]string, error)
}

type OSM struct {
	Client   *http.Client
	Endpoint string
}

type osmPlace struct {
	DisplayName string `json:"display_name"`
}

var (
	osmMu   sync.Mutex
	osmNext time.Time
)

func (o OSM) Suggest(ctx context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 3 {
		return []string{}, nil
	}
	if err := waitOSM(ctx); err != nil {
		return nil, err
	}
	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 4 * time.Second}
	}
	endpoint := o.Endpoint
	if endpoint == "" {
		endpoint = "https://nominatim.openstreetmap.org/search"
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	values := parsed.Query()
	values.Set("q", query)
	values.Set("format", "jsonv2")
	values.Set("limit", "6")
	parsed.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "CargoPrener/1.0 (educational group-buy demo)")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openstreetmap status %d", res.StatusCode)
	}
	var payload []osmPlace
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return osmLabels(payload), nil
}

func osmLabels(payload []osmPlace) []string {
	seen := map[string]struct{}{}
	items := make([]string, 0, len(payload))
	for _, item := range payload {
		label := strings.TrimSpace(item.DisplayName)
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

func waitOSM(ctx context.Context) error {
	osmMu.Lock()
	wait := time.Until(osmNext)
	if wait < 0 {
		wait = 0
	}
	osmNext = time.Now().Add(time.Second)
	osmMu.Unlock()
	if wait == 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
