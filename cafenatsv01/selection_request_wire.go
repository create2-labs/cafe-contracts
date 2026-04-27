package cafenatsv01

import (
	"errors"
	"fmt"
	"slices"
)

// Target posture strings align with CPM [github.com/create2-labs/cafe-cpm] vocabulary.CurrentPQPosture.
const (
	TargetPostureClassicalOnly = "classical_only"
	TargetPostureHybrid        = "hybrid"
	TargetPostureFullPQ        = "full_pq"
	TargetPostureUnknown       = "unknown"
)

// PolicySelectionRequestWire mirrors the JSON shape of the CPM PolicySelectionRequest
// for use on the wire without importing CPM. Callers may map it into the domain type.
type PolicySelectionRequestWire struct {
	TargetPosture             string   `json:"target_posture"`
	TargetChainIDs            []int64  `json:"target_chain_ids,omitempty"`
	RequireMultichain         bool     `json:"require_multichain"`
	AllowNewWallet            bool     `json:"allow_new_wallet"`
	AddressContinuityRequired bool     `json:"address_continuity_required"`
	KeyRotationRequired       bool     `json:"key_rotation_required"`
	RecoveryRequired          bool     `json:"recovery_required"`
	MinimumMaturity           int      `json:"minimum_maturity"`
	AllowResearch             bool     `json:"allow_research"`
	AllowedProviderModes      []string `json:"allowed_provider_modes,omitempty"`
	PreferredFamilies         []string `json:"preferred_families,omitempty"`
	PreferredProviders        []string `json:"preferred_providers,omitempty"`
	RequireBundlerAvailable   bool     `json:"require_bundler_available"`
	RequirePaymasterAvailable bool     `json:"require_paymaster_available"`
	ApprovalMode              string   `json:"approval_mode"`
}

// Normalize sets deterministic defaults (minimum maturity, approval mode, sorted chain ids).
func (p *PolicySelectionRequestWire) Normalize() {
	if p == nil {
		return
	}
	if p.MinimumMaturity == 0 {
		p.MinimumMaturity = 1
	}
	if p.ApprovalMode == "" {
		p.ApprovalMode = "manual"
	}
	p.TargetChainIDs = normalizeChainIDs(p.TargetChainIDs)
	p.AllowedProviderModes = normalizeStringSliceUniqueSorted(p.AllowedProviderModes)
	p.PreferredFamilies = dedupeStringPreserve(p.PreferredFamilies)
	p.PreferredProviders = dedupeStringPreserve(p.PreferredProviders)
}

// Validate checks the wire after optional Normalize.
func (p *PolicySelectionRequestWire) Validate() error {
	if p == nil {
		return errors.New("cafenatsv01 v0.1: selection_request is nil")
	}
	if p.TargetPosture == "" {
		return fmt.Errorf("%w: target_posture is required", ErrSelectionRequest)
	}
	if !isValidTargetPosture(p.TargetPosture) {
		return fmt.Errorf("%w: target_posture is invalid", ErrSelectionRequest)
	}
	for _, id := range p.TargetChainIDs {
		if id <= 0 {
			return fmt.Errorf("%w: target_chain_ids must be positive", ErrSelectionRequest)
		}
	}
	if p.RequireMultichain && len(p.TargetChainIDs) > 0 && len(p.TargetChainIDs) < 2 {
		return fmt.Errorf("%w: require_multichain needs at least two target_chain_ids when set", ErrSelectionRequest)
	}
	if p.MinimumMaturity < 1 || p.MinimumMaturity > 5 {
		return fmt.Errorf("%w: minimum_maturity must be 1..5", ErrSelectionRequest)
	}
	if p.ApprovalMode != "auto" && p.ApprovalMode != "manual" {
		return fmt.Errorf("%w: approval_mode must be auto or manual", ErrSelectionRequest)
	}
	for _, m := range p.AllowedProviderModes {
		if m != "third_party" && m != "user_managed" {
			return fmt.Errorf("%w: allowed_provider_modes invalid value %q", ErrSelectionRequest, m)
		}
	}
	return nil
}

func isValidTargetPosture(s string) bool {
	switch s {
	case TargetPostureClassicalOnly, TargetPostureHybrid, TargetPostureFullPQ, TargetPostureUnknown:
		return true
	default:
		return false
	}
}

func normalizeChainIDs(in []int64) []int64 {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, id := range in {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	slices.Sort(out)
	return out
}

func dedupeStringPreserve(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func normalizeStringSliceUniqueSorted(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		seen[s] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}
