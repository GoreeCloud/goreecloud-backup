package protection

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Policy defines the evidence requirements used to evaluate one protected
// system or dataset. Baseline GoreeCloud recovery evidence may be made stricter
// through freshness limits, but it may not be silently removed.
type Policy struct {
	ID                        string                `json:"id"`
	Requirements              []EvidenceRequirement `json:"requirements"`
	RestoreVerificationMaxAge time.Duration         `json:"restoreVerificationMaxAge"`
}

// EvidenceRequirement defines one required operational evidence check. A
// MaxAge of zero means that freshness is determined by the evidence producer's
// status rather than recalculated from an observation timestamp. Positive
// MaxAge values make time-derived freshness an additional gate.
type EvidenceRequirement struct {
	Kind   EvidenceKind  `json:"kind"`
	MaxAge time.Duration `json:"maxAge"`
}

// EvidenceObservation records the bounded status and observation time of one
// operational evidence check. It contains no backup contents or credentials.
type EvidenceObservation struct {
	Kind       EvidenceKind   `json:"kind"`
	Status     EvidenceStatus `json:"status"`
	ObservedAt time.Time      `json:"observedAt"`
}

// RestoreVerificationObservation records the current representative-restore
// verification status and when that result was observed.
type RestoreVerificationObservation struct {
	Status     EvidenceStatus `json:"status"`
	ObservedAt time.Time      `json:"observedAt"`
}

// ObservedAssessment is the freshness-aware input to policy evaluation.
type ObservedAssessment struct {
	Configured          bool                           `json:"configured"`
	BackupInProgress    bool                           `json:"backupInProgress"`
	Evidence            []EvidenceObservation          `json:"evidence"`
	RestoreVerification RestoreVerificationObservation `json:"restoreVerification"`
}

// BaselinePolicy returns the minimum GoreeCloud Backup evidence policy. It does
// not invent product-specific freshness durations; callers may set positive
// MaxAge values when the applicable policy defines them.
func BaselinePolicy() Policy {
	requirements := make([]EvidenceRequirement, 0, len(baselineRequiredEvidence))
	for _, kind := range baselineRequiredEvidence {
		requirements = append(requirements, EvidenceRequirement{Kind: kind})
	}

	return Policy{
		ID:           "baseline",
		Requirements: requirements,
	}
}

// Validate verifies that a policy is bounded, deterministic, and no weaker
// than the GoreeCloud Backup baseline recovery-evidence contract.
func (p Policy) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("policy ID must not be empty")
	}
	if p.RestoreVerificationMaxAge < 0 {
		return fmt.Errorf("restore verification max age must not be negative")
	}

	seen := make(map[EvidenceKind]struct{}, len(p.Requirements))
	for _, requirement := range p.Requirements {
		if !requirement.Kind.valid() {
			return fmt.Errorf("invalid policy evidence kind %q", requirement.Kind)
		}
		if requirement.MaxAge < 0 {
			return fmt.Errorf("max age for evidence %q must not be negative", requirement.Kind)
		}
		if _, ok := seen[requirement.Kind]; ok {
			return fmt.Errorf("duplicate policy evidence requirement %q", requirement.Kind)
		}
		seen[requirement.Kind] = struct{}{}
	}

	var missing []EvidenceKind
	for _, kind := range baselineRequiredEvidence {
		if _, ok := seen[kind]; !ok {
			missing = append(missing, kind)
		}
	}
	if len(missing) > 0 {
		sortEvidenceKinds(missing)
		return fmt.Errorf("policy cannot remove baseline evidence requirements: %v", missing)
	}

	return nil
}

// EvaluateObserved evaluates freshness-aware evidence at an explicit point in
// time. It never reads the system clock, keeping state calculation deterministic
// for APIs, persisted evidence replay, tests, and incident review.
func EvaluateObserved(policy Policy, observed ObservedAssessment, evaluatedAt time.Time) (Evaluation, error) {
	if err := policy.Validate(); err != nil {
		return Evaluation{}, err
	}
	if evaluatedAt.IsZero() {
		return Evaluation{}, fmt.Errorf("evaluation time must not be zero")
	}

	requirements := make(map[EvidenceKind]EvidenceRequirement, len(policy.Requirements))
	for _, requirement := range policy.Requirements {
		requirements[requirement.Kind] = requirement
	}

	assessment := Assessment{
		Configured:       observed.Configured,
		BackupInProgress: observed.BackupInProgress,
	}

	seen := make(map[EvidenceKind]struct{}, len(observed.Evidence))
	for _, observation := range observed.Evidence {
		if !observation.Kind.valid() {
			return Evaluation{}, fmt.Errorf("invalid observed evidence kind %q", observation.Kind)
		}
		if _, ok := seen[observation.Kind]; ok {
			return Evaluation{}, fmt.Errorf("duplicate observed evidence kind %q", observation.Kind)
		}
		seen[observation.Kind] = struct{}{}

		status := normalizeEvidenceStatus(observation.Status)
		if !status.valid() {
			return Evaluation{}, fmt.Errorf("invalid observed status %q for evidence %q", observation.Status, observation.Kind)
		}
		if !observation.ObservedAt.IsZero() && observation.ObservedAt.After(evaluatedAt) {
			return Evaluation{}, fmt.Errorf("evidence %q observation time is after evaluation time", observation.Kind)
		}

		requirement, required := requirements[observation.Kind]
		if required && status == EvidencePassing && requirement.MaxAge > 0 {
			if observation.ObservedAt.IsZero() || evaluatedAt.Sub(observation.ObservedAt) > requirement.MaxAge {
				status = EvidenceStale
			}
		}

		assessment.Evidence = append(assessment.Evidence, EvidenceItem{
			Kind:     observation.Kind,
			Status:   status,
			Required: required,
		})
	}

	restoreStatus := normalizeEvidenceStatus(observed.RestoreVerification.Status)
	if !restoreStatus.valid() {
		return Evaluation{}, fmt.Errorf("invalid restore verification status %q", observed.RestoreVerification.Status)
	}
	if !observed.RestoreVerification.ObservedAt.IsZero() && observed.RestoreVerification.ObservedAt.After(evaluatedAt) {
		return Evaluation{}, fmt.Errorf("restore verification observation time is after evaluation time")
	}
	if restoreStatus == EvidencePassing && policy.RestoreVerificationMaxAge > 0 {
		if observed.RestoreVerification.ObservedAt.IsZero() || evaluatedAt.Sub(observed.RestoreVerification.ObservedAt) > policy.RestoreVerificationMaxAge {
			restoreStatus = EvidenceStale
		}
	}
	assessment.RestoreVerification = restoreStatus

	return Evaluate(assessment)
}

// RequirementKinds returns the policy's required evidence kinds in stable
// lexical order. The result is a new slice and may be modified by the caller.
func (p Policy) RequirementKinds() []EvidenceKind {
	kinds := make([]EvidenceKind, 0, len(p.Requirements))
	for _, requirement := range p.Requirements {
		kinds = append(kinds, requirement.Kind)
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
	return kinds
}
