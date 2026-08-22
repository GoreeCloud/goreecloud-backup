package protection_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/goreecloud/protection"
)

func passingOperationalEvidence() []protection.EvidenceItem {
	return []protection.EvidenceItem{
		{Kind: protection.EvidenceRepositoryAvailable, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceCredentialsRecoverable, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceRecoveryPointAvailable, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceBackupCurrent, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceIntegrity, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceScope, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceApplicationConsistency, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceRetention, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceMaintenance, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceMonitoring, Status: protection.EvidencePassing, Required: true},
		{Kind: protection.EvidenceNotification, Status: protection.EvidencePassing, Required: true},
	}
}

func TestEvaluateProtectionState(t *testing.T) {
	tests := []struct {
		name       string
		assessment protection.Assessment
		want       protection.State
		wantReason protection.ReasonCode
	}{
		{
			name: "not configured is unprotected",
			assessment: protection.Assessment{
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateUnprotected,
			wantReason: protection.ReasonNotConfigured,
		},
		{
			name: "configured with no evidence remains configured",
			assessment: protection.Assessment{
				Configured:          true,
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateConfigured,
			wantReason: protection.ReasonNoRequiredEvidence,
		},
		{
			name: "active backup is backing up",
			assessment: protection.Assessment{
				Configured:          true,
				BackupInProgress:    true,
				Evidence:            passingOperationalEvidence(),
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateBackingUp,
			wantReason: protection.ReasonBackupInProgress,
		},
		{
			name: "missing required evidence remains configured",
			assessment: protection.Assessment{
				Configured: true,
				Evidence: []protection.EvidenceItem{
					{Kind: protection.EvidenceRecoveryPointAvailable, Status: protection.EvidencePassing, Required: true},
					{Kind: protection.EvidenceIntegrity, Status: protection.EvidenceUnknown, Required: true},
				},
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateConfigured,
			wantReason: protection.ReasonRequiredEvidenceMissing,
		},
		{
			name: "snapshot success alone is not protected",
			assessment: protection.Assessment{
				Configured: true,
				Evidence: []protection.EvidenceItem{
					{Kind: protection.EvidenceRecoveryPointAvailable, Status: protection.EvidencePassing, Required: true},
					{Kind: protection.EvidenceIntegrity, Status: protection.EvidenceUnknown, Required: true},
				},
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateConfigured,
			wantReason: protection.ReasonRequiredEvidenceMissing,
		},
		{
			name: "passing operational evidence is protected",
			assessment: protection.Assessment{
				Configured:          true,
				Evidence:            passingOperationalEvidence(),
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateProtected,
			wantReason: protection.ReasonOperationalEvidencePassing,
		},
		{
			name: "restore verification produces restore verified",
			assessment: protection.Assessment{
				Configured:          true,
				Evidence:            passingOperationalEvidence(),
				RestoreVerification: protection.EvidencePassing,
			},
			want:       protection.StateRestoreVerified,
			wantReason: protection.ReasonRestoreVerificationPassing,
		},
		{
			name: "required operational failure is degraded",
			assessment: protection.Assessment{
				Configured: true,
				Evidence: []protection.EvidenceItem{
					{Kind: protection.EvidenceRepositoryAvailable, Status: protection.EvidenceFailing, Required: true},
				},
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateDegraded,
			wantReason: protection.ReasonRequiredEvidenceFailed,
		},
		{
			name: "stale restore verification is degraded",
			assessment: protection.Assessment{
				Configured:          true,
				Evidence:            passingOperationalEvidence(),
				RestoreVerification: protection.EvidenceStale,
			},
			want:       protection.StateDegraded,
			wantReason: protection.ReasonRestoreVerificationStale,
		},
		{
			name: "failure wins over backup in progress",
			assessment: protection.Assessment{
				Configured:       true,
				BackupInProgress: true,
				Evidence: []protection.EvidenceItem{
					{Kind: protection.EvidenceRepositoryAvailable, Status: protection.EvidenceFailing, Required: true},
				},
				RestoreVerification: protection.EvidenceUnknown,
			},
			want:       protection.StateDegraded,
			wantReason: protection.ReasonRequiredEvidenceFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := protection.Evaluate(tt.assessment)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.State)
			require.Contains(t, got.Reasons, tt.wantReason)
		})
	}
}

func TestEvaluateRejectsInvalidEvidence(t *testing.T) {
	t.Run("duplicate evidence kind", func(t *testing.T) {
		_, err := protection.Evaluate(protection.Assessment{
			Configured: true,
			Evidence: []protection.EvidenceItem{
				{Kind: protection.EvidenceIntegrity, Status: protection.EvidencePassing, Required: true},
				{Kind: protection.EvidenceIntegrity, Status: protection.EvidencePassing, Required: true},
			},
			RestoreVerification: protection.EvidenceUnknown,
		})
		require.ErrorContains(t, err, "duplicate evidence kind")
	})

	t.Run("required evidence cannot be not applicable", func(t *testing.T) {
		_, err := protection.Evaluate(protection.Assessment{
			Configured: true,
			Evidence: []protection.EvidenceItem{
				{Kind: protection.EvidenceIntegrity, Status: protection.EvidenceNotApplicable, Required: true},
			},
			RestoreVerification: protection.EvidenceUnknown,
		})
		require.ErrorContains(t, err, "cannot be not applicable")
	})
}

func TestEvaluationSortsEvidenceForStableAPIResults(t *testing.T) {
	got, err := protection.Evaluate(protection.Assessment{
		Configured: true,
		Evidence: []protection.EvidenceItem{
			{Kind: protection.EvidenceScope, Status: protection.EvidenceFailing, Required: true},
			{Kind: protection.EvidenceBackupCurrent, Status: protection.EvidenceFailing, Required: true},
		},
		RestoreVerification: protection.EvidenceUnknown,
	})
	require.NoError(t, err)
	require.Equal(t, []protection.EvidenceKind{
		protection.EvidenceBackupCurrent,
		protection.EvidenceScope,
	}, got.Failed)
}

func TestRecoveryEvidenceValidate(t *testing.T) {
	now := time.Date(2026, time.August, 22, 18, 0, 0, 0, time.UTC)

	valid := protection.RecoveryEvidence{
		DatasetID:       "family-photos",
		RepositoryID:    "repo-primary",
		RecoveryPointID: "snapshot-123",
		ObservedAt:      now,
		SoftwareVersion: "development",
		BackupStatus:    protection.EvidencePassing,
		IntegrityStatus: protection.EvidencePassing,
		RestoreTest: &protection.RestoreTestEvidence{
			Type:        protection.VerificationRepresentativeRestore,
			Status:      protection.EvidencePassing,
			CompletedAt: now,
			Checks: []protection.ValidationCheck{
				protection.ValidationContentHash,
				protection.ValidationMetadata,
				protection.ValidationPermissions,
			},
			FailureCategory: protection.FailureNone,
		},
	}

	require.NoError(t, valid.Validate())

	invalidFailure := valid
	invalidFailure.RestoreTest = &protection.RestoreTestEvidence{
		Type:            protection.VerificationFileSample,
		Status:          protection.EvidenceFailing,
		CompletedAt:     now,
		FailureCategory: protection.FailureNone,
	}
	require.ErrorContains(t, invalidFailure.Validate(), "must have a failure category")

	invalidDuplicateCheck := valid
	invalidDuplicateCheck.RestoreTest = &protection.RestoreTestEvidence{
		Type:        protection.VerificationFileSample,
		Status:      protection.EvidencePassing,
		CompletedAt: now,
		Checks: []protection.ValidationCheck{
			protection.ValidationContentHash,
			protection.ValidationContentHash,
		},
		FailureCategory: protection.FailureNone,
	}
	require.ErrorContains(t, invalidDuplicateCheck.Validate(), "duplicate validation check")
}
