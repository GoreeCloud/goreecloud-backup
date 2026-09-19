package protection

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var errInvalidProtectionPolicy = errors.New("invalid protection policy")

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
		requirements = append(requirements, EvidenceRequirement{Kind: kind, MaxAge: 0})
	}

	return Policy{
		ID:                        "baseline",
		Requirements:              requirements,
		RestoreVerificationMaxAge: 0,
	}
}

// Validate verifies that a policy is bounded, deterministic, and no weaker
// than the GoreeCloud Backup baseline recovery-evidence contract.
func (p Policy) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errPolicyIDEmpty
	}

	if p.RestoreVerificationMaxAge < 0 {
		return errRestoreVerificationMaxAgeNegative
	}

	seen := make(map[EvidenceKind]struct{}, len(p.Requirements))
	for _, requirement := range p.Requirements {
		if !requirement.Kind.valid() {
			return fmt.Errorf("%w: invalid policy evidence kind %q", errInvalidProtectionPolicy, requirement.Kind)
		}

		if requirement.MaxAge < 0 {
			return fmt.Errorf("%w: max age for evidence %q must not be negative", errInvalidProtectionPolicy, requirement.Kind)
		}

		if _, ok := seen[requirement.Kind]; ok {
			return fmt.Errorf("%w: duplicate policy evidence requirement %q", errInvalidProtectionPolicy, requirement.Kind)
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
		return fmt.Errorf("%w: policy cannot remove baseline evidence requirements: %v", errInvalidProtectionPolicy, missing)
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
		return Evaluation{}, errEvaluationTimeZero
	}

	requirements := make(map[EvidenceKind]EvidenceRequirement, len(policy.Requirements))
	for _, requirement := range policy.Requirements {
		requirements[requirement.Kind] = requirement
	}

	assessment := Assessment{
		Configured:          observed.Configured,
		BackupInProgress:    observed.BackupInProgress,
		Evidence:            nil,
		RestoreVerification: "",
	}

	seen := make(map[EvidenceKind]struct{}, len(observed.Evidence))
	for _, observation := range observed.Evidence {
		if !observation.Kind.valid() {
			return Evaluation{}, fmt.Errorf("%w: invalid observed evidence kind %q", errInvalidProtectionPolicy, observation.Kind)
		}

		if _, ok := seen[observation.Kind]; ok {
			return Evaluation{}, fmt.Errorf("%w: duplicate observed evidence kind %q", errInvalidProtectionPolicy, observation.Kind)
		}

		seen[observation.Kind] = struct{}{}

		status := normalizeEvidenceStatus(observation.Status)
		if !status.valid() {
			return Evaluation{}, fmt.Errorf("%w: invalid observed status %q for evidence %q", errInvalidProtectionPolicy, observation.Status, observation.Kind)
		}

		if !observation.ObservedAt.IsZero() && observation.ObservedAt.After(evaluatedAt) {
			return Evaluation{}, fmt.Errorf("%w: evidence %q observation time is after evaluation time", errInvalidProtectionPolicy, observation.Kind)
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
		return Evaluation{}, fmt.Errorf("%w: invalid restore verification status %q", errInvalidProtectionPolicy, observed.RestoreVerification.Status)
	}

	if !observed.RestoreVerification.ObservedAt.IsZero() && observed.RestoreVerification.ObservedAt.After(evaluatedAt) {
		return Evaluation{}, errRestoreVerificationObservedAfterEvaluation
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

	slices.Sort(kinds)

	return kinds
}
