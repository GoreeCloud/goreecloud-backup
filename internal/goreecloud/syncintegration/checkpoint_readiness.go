package syncintegration

// ReadyForSubmission reports whether this Backup-owned status is both bound to
// the exact accepted checkpoint submission and strong enough for the source-
// level protected-change predicate.
//
// Runtime orchestration should prefer this method when deciding whether a
// concrete pre-change or pre-migration request may proceed. A controlling
// Backup/Everkeep policy may still impose additional requirements.
func (s CheckpointStatus) ReadyForSubmission(submission CheckpointSubmission) bool {
	return s.ValidateForSubmission(submission) == nil &&
		s.State == CheckpointStateCompleted &&
		s.RecoveryPointUsable &&
		s.IntegrityVerified
}
