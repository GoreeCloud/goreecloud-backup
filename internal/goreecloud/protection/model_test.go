package protection_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/goreecloud/protection"
)

func passingOperationalEvidence() []protection.EvidenceItem {
	return []protection.EvidenceItem{
		{Kind: protection.EvidenceRepositoryAvailable, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceCredentialsRecoverable, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceRecoveryPointAvailable, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceBackupCurrent, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceIntegrity, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceScope, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceApplicationConsistency, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceRetention, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceMaintenance, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceMonitoring, Status: protection.EvidencePassing},
		{Kind: protection.EvidenceNotification, Status: protection.EvidencePassing},
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
			name:       "zero value is unprotected",
			assessment: protection.Assessment{},
			want:       protection.StateUnprotected,
			wantReason: protection.ReasonNotConfigured,
		},
		{
			name: "configured with no evidence remains configured",
			assessment: protection.Assessment{
				Configured: true,
			},
			want:       protection.StateConfigured,
			wantReason: protection.ReasonRequiredEvidenceMissing,
		},
		{
			name: "active healthy backup is backing up",
			assessment: protection.Assessment{
				Configured:       true,
				BackupInProgress: true,
				Evidence:         passingOperationalEvidence(),
			},
			want:       protection.StateBackingUp,
			wantReason: protection.ReasonBackupInProgress,
		},
		{
			name: "snapshot success alone remains configured",
			assessment: protection.Assessment{
				Configured: true,
				Evidence: []protection.EvidenceItem{
					{Kind: protection.EvidenceRecoveryPointAvailable, Status: protection.EvidencePassing},
				},
			},
			want:       protection.StateConfigured,
			wantReason: protection.ReasonRequiredEvidenceMissing,
		},
		{
			name: "passing operational evidence is protected",
			assessment: protection.Assessment{
				Configured: true,
				Evidence:   passingOperationalEvidence(),
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
					{Kind: protection.EvidenceRepositoryAvailable, Status: protection.EvidenceFailing},
				},
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
					{Kind: protection.EvidenceRepositoryAvailable, Status: protection.EvidenceFailing},
				},
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

func TestSnapshotSuccessCannotSatisfyBaselineEvidence(t *testing.T) {
	got, err := protection.Evaluate(protection.Assessment{
		Configured: true,
		Evidence: []protection.EvidenceItem{
			{Kind: protection.EvidenceRecoveryPointAvailable, Status: protection.EvidencePassing},
		},
	})
	require.NoError(t, err)
	require.Equal(t, protection.StateConfigured, got.State)
	require.Contains(t, got.Missing, protection.EvidenceRepositoryAvailable)
	require.Contains(t, got.Missing, protection.EvidenceIntegrity)
	require.Contains(t, got.Missing, protection.EvidenceMonitoring)
	require.NotContains(t, got.Missing, protection.EvidenceRecoveryPointAvailable)
}

func TestBaselineRequiredEvidenceReturnsCopy(t *testing.T) {
	first := protection.BaselineRequiredEvidence()
	require.NotEmpty(t, first)
	first[0] = protection.EvidenceIntegrity

	second := protection.BaselineRequiredEvidence()
	require.Equal(t, protection.EvidenceRepositoryAvailable, second[0])
}

func TestEvaluateRejectsInvalidEvidence(t *testing.T) {
	t.Run("duplicate evidence kind", func(t *testing.T) {
		_, err := protection.Evaluate(protection.Assessment{
			Configured: true,
			Evidence: []protection.EvidenceItem{
				{Kind: protection.EvidenceIntegrity, Status: protection.EvidencePassing},
				{Kind: protection.EvidenceIntegrity, Status: protection.EvidencePassing},
			},
		})
		require.ErrorContains(t, err, "duplicate evidence kind")
	})

	t.Run("baseline evidence cannot be not applicable", func(t *testing.T) {
		_, err := protection.Evaluate(protection.Assessment{
			Configured: true,
			Evidence: []protection.EvidenceItem{
				{Kind: protection.EvidenceIntegrity, Status: protection.EvidenceNotApplicable},
			},
		})
		require.ErrorContains(t, err, "cannot be not applicable")
	})

	t.Run("unknown evidence kind", func(t *testing.T) {
		_, err := protection.Evaluate(protection.Assessment{
			Configured: true,
			Evidence: []protection.EvidenceItem{
				{Kind: protection.EvidenceKind("arbitrary_check"), Status: protection.EvidencePassing},
			},
		})
		require.ErrorContains(t, err, "invalid evidence kind")
	})
}

func TestEvaluationSortsEvidenceForStableAPIResults(t *testing.T) {
	evidence := passingOperationalEvidence()
	for i := range evidence {
		switch evidence[i].Kind {
		case protection.EvidenceScope, protection.EvidenceBackupCurrent:
			evidence[i].Status = protection.EvidenceFailing
		}
	}

	got, err := protection.Evaluate(protection.Assessment{
		Configured: true,
		Evidence:   evidence,
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
