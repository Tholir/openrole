// Package api provides HTTP clients for external services.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Cache entry with TTL tracking.
type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// Open5eCache provides an in-memory TTL cache for API responses.
type Open5eCache struct {
	entries map[string]cacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

// NewOpen5eCache creates a new cache with the specified TTL.
func NewOpen5eCache(ttl time.Duration) *Open5eCache {
	return &Open5eCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// CacheGet retrieves a cached value if it exists and is not expired.
// Returns nil if the key is not found or has expired.
func (c *Open5eCache) CacheGet(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

// CacheSet stores a value in the cache with the configured TTL.
func (c *Open5eCache) CacheSet(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// cacheKey builds a cache key from endpoint and parameters.
func cacheKey(endpoint string, params ...string) string {
	key := endpoint
	for _, p := range params {
		key += "/" + p
	}
	return key
}

// Open5eClient is an HTTP client for the open5e.com API.
type Open5eClient struct {
	baseURL string
	client  *http.Client
	cache   *Open5eCache
}

// NewOpen5eClient creates a new Open5e API client with a 24-hour TTL cache.
func NewOpen5eClient() *Open5eClient {
	return &Open5eClient{
		baseURL: "https://api.open5e.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: NewOpen5eCache(24 * time.Hour),
	}
}

// NewOpen5eClientWithCache creates a new Open5e API client with a custom cache.
func NewOpen5eClientWithCache(cache *Open5eCache) *Open5eClient {
	return &Open5eClient{
		baseURL: "https://api.open5e.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: cache,
	}
}

// Spell represents a D&D 5e spell from open5e.
type Spell struct {
	Slug           string   `json:"slug"`
	Name           string   `json:"name"`
	Desc           string   `json:"desc"`
	HigherLevel    []string `json:"higher_level"`
	Range          string   `json:"range"`
	Components     []string `json:"components"`
	Material       string   `json:"material"`
	Ritual         bool     `json:"ritual"`
	RitualStr      string   `json:"ritual_str"`
	Concentration bool     `json:"concentration"`
	ConcentrationStr string `json:"concentration_str"`
	CastingTime    string   `json:"casting_time"`
	Duration       string   `json:"duration"`
	Level          int      `json:"level"`
	LevelStr       string   `json:"level_str"`
	School         string   `json:"school"`
	Classes        []string `json:"classes"`
	Subclasses     []string `json:"subclasses"`
	URL            string   `json:"url"`
}

// SpellListResponse represents the paginated spell list response.
type SpellListResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []Spell `json:"results"`
}

// Monster represents a D&D 5e creature from open5e.
type Monster struct {
	Slug            string  `json:"slug"`
	Name            string  `json:"name"`
	Size            string  `json:"size"`
	Type            string  `json:"type"`
	Alignment       string  `json:"alignment"`
	ArmorClass      []int   `json:"armor_class"`
	HitPoints       int     `json:"hit_points"`
	HitDice         string  `json:"hit_dice"`
	Speed           map[string]any `json:"speed"`
	Strength        int     `json:"strength"`
	Dexterity       int     `json:"dexterity"`
	Constitution    int     `json:"constitution"`
	Intelligence    int     `json:"intelligence"`
	Wisdom          int     `json:"wisdom"`
	Charisma        int     `json:"charisma"`
	Proficiencies   []any   `json:"proficiencies"`
	DamageVulnerabilities []string `json:"damage_vulnerabilities"`
	DamageResistances     []string `json:"damage_resistances"`
	DamageImmunities      []string `json:"damage_immunities"`
	ConditionImmunities   []string `json:"condition_immunities"`
	Senses          map[string]any `json:"senses"`
	Languages       string  `json:"languages"`
	ChallengeRating string  `json:"challenge_rating"`
	XP              int     `json:"xp"`
	URL            string   `json:"url"`
	Actions         []any   `json:"actions"`
	Reactions       []any   `json:"reactions"`
	LegendaryActions []any  `json:"legendary_actions"`
}

// MonsterListResponse represents the paginated monster list response.
type MonsterListResponse struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []Monster `json:"results"`
}

// GetSpell fetches a single spell by its slug.
func (c *Open5eClient) GetSpell(ctx context.Context, slug string) (*Spell, error) {
	key := cacheKey("spell", slug)
	if data, ok := c.cache.CacheGet(key); ok {
		var spell Spell
		if err := json.Unmarshal(data, &spell); err == nil {
			return &spell, nil
		}
	}

	url := fmt.Sprintf("%s/v1/spells/%s/", c.baseURL, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spell: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var spell Spell
	if err := json.NewDecoder(resp.Body).Decode(&spell); err != nil {
		return nil, fmt.Errorf("failed to decode spell: %w", err)
	}

	if data, err := json.Marshal(spell); err == nil {
		c.cache.CacheSet(key, data)
	}

	return &spell, nil
}

// ListSpells fetches a paginated list of spells.
func (c *Open5eClient) ListSpells(ctx context.Context, page int) (*SpellListResponse, error) {
	url := fmt.Sprintf("%s/v1/spells/?page=%d", c.baseURL, page)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spells: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result SpellListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode spells: %w", err)
	}
	return &result, nil
}

// SearchSpells searches for spells by name.
func (c *Open5eClient) SearchSpells(ctx context.Context, name string) ([]Spell, error) {
	url := fmt.Sprintf("%s/v1/spells/?search=%s", c.baseURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search spells: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result SpellListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode spells: %w", err)
	}
	return result.Results, nil
}

// GetMonster fetches a single monster by its slug.
func (c *Open5eClient) GetMonster(ctx context.Context, slug string) (*Monster, error) {
	url := fmt.Sprintf("%s/v1/monsters/%s/", c.baseURL, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var monster Monster
	if err := json.NewDecoder(resp.Body).Decode(&monster); err != nil {
		return nil, fmt.Errorf("failed to decode monster: %w", err)
	}
	return &monster, nil
}

// ListMonsters fetches a paginated list of monsters.
func (c *Open5eClient) ListMonsters(ctx context.Context, page int) (*MonsterListResponse, error) {
	url := fmt.Sprintf("%s/v1/monsters/?page=%d", c.baseURL, page)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monsters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result MonsterListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode monsters: %w", err)
	}
	return &result, nil
}

// SearchMonsters searches for monsters by name.
func (c *Open5eClient) SearchMonsters(ctx context.Context, name string) ([]Monster, error) {
	url := fmt.Sprintf("%s/v1/monsters/?search=%s", c.baseURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search monsters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result MonsterListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode monsters: %w", err)
	}
	return result.Results, nil
}

// GetDocument fetches SRD document metadata.
func (c *Open5eClient) GetDocument(ctx context.Context, slug string) (map[string]any, error) {
	url := fmt.Sprintf("%s/v1/documents/%s/", c.baseURL, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}
	return result, nil
}