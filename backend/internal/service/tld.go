package service

import (
	"context"
	"sort"
	"strings"

	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
)

// TLDService handles TLD operations
type TLDService struct {
	client *NamecheapClient
	cache  []model.TLD
}

// NewTLDService creates a new TLD service
func NewTLDService(client *NamecheapClient) *TLDService {
	return &TLDService{
		client: client,
	}
}

// GetTLDs retrieves the list of all TLDs
func (s *TLDService) GetTLDs(ctx context.Context) ([]model.TLD, error) {
	if s.cache != nil {
		return s.cache, nil
	}

	tldsData, err := s.client.GetTldList(ctx)
	if err != nil {
		return nil, err
	}

	tlds := make([]model.TLD, 0, len(tldsData))
	seen := make(map[string]bool)

	for _, t := range tldsData {
		name := strings.ToLower(strings.TrimSpace(t["Name"]))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		isApiRegisterable := strings.ToLower(strings.TrimSpace(t["IsApiRegisterable"])) == "true"
		isActive := strings.ToLower(strings.TrimSpace(t["IsActive"])) == "true"

		tlds = append(tlds, model.TLD{
			Name:              name,
			IsApiRegisterable: isApiRegisterable,
			IsActive:          isActive,
		})
	}

	// Sort by name
	sort.Slice(tlds, func(i, j int) bool {
		return tlds[i].Name < tlds[j].Name
	})

	s.cache = tlds
	return tlds, nil
}

// GetTLDsByMode filters TLDs by mode
func (s *TLDService) GetTLDsByMode(ctx context.Context, mode model.TldMode) ([]model.TLD, error) {
	tlds, err := s.GetTLDs(ctx)
	if err != nil {
		return nil, err
	}

	if mode == model.TldModeAll {
		return tlds, nil
	}

	var filtered []model.TLD
	for _, tld := range tlds {
		if mode == model.TldModeApiRegisterableOnly && tld.IsApiRegisterable {
			filtered = append(filtered, tld)
		}
	}

	return filtered, nil
}
