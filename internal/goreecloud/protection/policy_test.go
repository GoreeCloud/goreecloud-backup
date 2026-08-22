package protection_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/goreecloud/protection"
)

func observationsAt(observedAt time.Time) []protection.EvidenceObservation {
	items := passingOperationalEvidence()
	out := make([]protection.EvidenceObservation, 0, len(items))
	for _, item := range items {
		out = append(out, protection.EvidenceObservation{
			Kind:       item.Kind,
			Status:     item.Status,
			ObservedAt: observedAt,
		})
	}
	return out
}

func policyWithMaxAge(t *testing.T, kind protection.EvidenceKind, maxAge time.Duration) protection.Policy {
	t.Helper()

	policy := protection.BaselinePolicy()
	policy.ID = "test-policy"
	found := false
	for i := range policy.Requirements {
		if policy.Requirements[i].Kind == kind {
			policy.Requirements[i].MaxAge = maxAge
			found = true
			break
		}
	}
	require.True(t, found, "expected baseline requirement %q", kind)
	return policy
}

func TestBaselinePolicyValid(t *testing.T) {
	policy := protection.BaselinePolicy()
	require.NoError(t, policy.Validate())
	require.ElementsMatch(t, protection.BaselineRequiredEvidence(), policy.RequirementKinds())
}

func TestPolicyCannotRemoveBaselineEvidence(t *testing.T) {
	policy := protection.BaselinePolicy()
	policy.Requirements = policy.Requirements[1:]

	err := policy.Validate()
	require.ErrorContains(t, err, "cannot remove baseline evidence requirements")
	require.ErrorContains(t, err, string(protection.EvidenceRepositoryAvailable))
}

func TestPolicyRejectsInvalidFreshness(t *testing.T) {
	t.Run("negative evidence max age", func(t *testing.T) {
		policy := policyWithMaxAge(t, protection.EvidenceIntegrity, -time.Minute)
		require.ErrorContains(t, policy.Validate(), "must not be negative")
	})

	t.Run("negative restore max age", func(t *testing.T) {
		policy := protection.BaselinePolicy()
		policy.RestoreVerificationMaxAge = -time.Minute
		require.ErrorContains(t, policy.Validate(), "must not be negative")
	})

	t.Run("duplicate requirement", func(t *testing.T) {
		policy := protection.BaselinePolicy()
		policy.Requirements = append(policy.Requirements, policy.Requirements[0])
		require.ErrorContains(t, policy.Validate(), "duplicate policy evidence requirement")
	})
}

func TestEvaluateObservedFreshness(t *testing.T) {
	evaluatedAt := time.Date(2026, time.August, 22, 18, 0, 0, 0, time.UTC)

	t.Run("fresh required evidence remains protected", func(t *testing.T) {
		policy := policyWithMaxAge(t, protection.EvidenceIntegrity, time.Hour)
		got, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
			Configured: true,
			Evidence:   observationsAt(evaluatedAt.Add(-30 * time.Minute)),
		}, evaluatedAt)
		require.NoError(t, err)
		require.Equal(t, protection.StateProtected, got.State)
	})

	t.Run("old passing evidence becomes stale and degraded", func(t *testing.T) {
		policy := policyWithMaxAge(t, protection.EvidenceIntegrity, time.Hour)
		got, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
			Configured: true,
			Evidence:   observationsAt(evaluatedAt.Add(-2 * time.Hour)),
		}, evaluatedAt)
		require.NoError(t, err)
		require.Equal(t, protection.StateDegraded, got.State)
		require.Contains(t, got.Stale, protection.EvidenceIntegrity)
		require.Contains(t, got.Reasons, protection.ReasonRequiredEvidenceStale)
	})

	t.Run("passing evidence without timestamp becomes stale when policy requires freshness", func(t *testing.T) {
		policy := policyWithMaxAge(t, protection.EvidenceIntegrity, time.Hour)
		observations := observationsAt(evaluatedAt)
		for i := range observations {
			if observations[i].Kind == protection.EvidenceIntegrity {
				observations[i].ObservedAt = time.Time{}
			}
		}

		got, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
			Configured: true,
			Evidence:   observations,
		}, evaluatedAt)
		require.NoError(t, err)
		require.Equal(t, protection.StateDegraded, got.State)
		require.Contains(t, got.Stale, protection.EvidenceIntegrity)
	})

	t.Run("producer supplied stale status remains degraded without max age", func(t *testing.T) {
		policy := protection.BaselinePolicy()
		observations := observationsAt(evaluatedAt)
		for i := range observations {
			if observations[i].Kind == protection.EvidenceMonitoring {
				observations[i].Status = protection.EvidenceStale
			}
		}

		got, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
			Configured: true,
			Evidence:   observations,
		}, evaluatedAt)
		require.NoError(t, err)
		require.Equal(t, protection.StateDegraded, got.State)
		require.Contains(t, got.Stale, protection.EvidenceMonitoring)
	})
}

func TestEvaluateObservedRestoreVerificationFreshness(t *testing.T) {
	evaluatedAt := time.Date(2026, time.August, 22, 18, 0, 0, 0, time.UTC)
	policy := protection.BaselinePolicy()
	policy.RestoreVerificationMaxAge = 24 * time.Hour

	fresh, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
		Configured: true,
		Evidence:   observationsAt(evaluatedAt),
		RestoreVerification: protection.RestoreVerificationObservation{
			Status:     protection.EvidencePassing,
			ObservedAt: evaluatedAt.Add(-12 * time.Hour),
		},
	}, evaluatedAt)
	require.NoError(t, err)
	require.Equal(t, protection.StateRestoreVerified, fresh.State)

	stale, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
		Configured: true,
		Evidence:   observationsAt(evaluatedAt),
		RestoreVerification: protection.RestoreVerificationObservation{
			Status:     protection.EvidencePassing,
			ObservedAt: evaluatedAt.Add(-48 * time.Hour),
		},
	}, evaluatedAt)
	require.NoError(t, err)
	require.Equal(t, protection.StateDegraded, stale.State)
	require.Contains(t, stale.Reasons, protection.ReasonRestoreVerificationStale)
}

func TestEvaluateObservedRejectsInvalidTimeInputs(t *testing.T) {
	policy := protection.BaselinePolicy()
	evaluatedAt := time.Date(2026, time.August, 22, 18, 0, 0, 0, time.UTC)

	t.Run("zero evaluation time", func(t *testing.T) {
		_, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{}, time.Time{})
		require.ErrorContains(t, err, "evaluation time must not be zero")
	})

	t.Run("future evidence observation", func(t *testing.T) {
		_, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
			Configured: true,
			Evidence: []protection.EvidenceObservation{
				{
					Kind:       protection.EvidenceIntegrity,
					Status:     protection.EvidencePassing,
					ObservedAt: evaluatedAt.Add(time.Minute),
				},
			},
		}, evaluatedAt)
		require.ErrorContains(t, err, "observation time is after evaluation time")
	})

	t.Run("future restore verification observation", func(t *testing.T) {
		_, err := protection.EvaluateObserved(policy, protection.ObservedAssessment{
			RestoreVerification: protection.RestoreVerificationObservation{
				Status:     protection.EvidencePassing,
				ObservedAt: evaluatedAt.Add(time.Minute),
			},
		}, evaluatedAt)
		require.ErrorContains(t, err, "restore verification observation time is after evaluation time")
	})
}
