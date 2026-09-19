package syncintegration

import "time"

// ReadyForSubmission reports whether this Backup-owned status is both bound to
// the exact accepted checkpoint submission and strong enough for the source-
// level protected-change predicate.
//
// Runtime orchestration may use this lower-level predicate when no freshness
// policy applies. When a controlling Backup/Everkeep policy defines a maximum
// checkpoint age, prefer ReadyForSubmissionWithin so stale evidence fails
// closed instead of being treated as current merely because it remains valid.
func (s CheckpointStatus) ReadyForSubmission(submission CheckpointSubmission) bool {
	return s.ValidateForSubmission(submission) == nil &&
		s.State == CheckpointStateCompleted &&
		s.RecoveryPointUsable &&
		s.IntegrityVerified
}

// ReadyForSubmissionWithin adds a caller-supplied freshness boundary to the
// exact-binding and recovery-evidence requirements of ReadyForSubmission.
// Backup/Everkeep policy owns maxAge; GoreeCloud Sync does not choose or extend
// it through this contract.
//
// Zero/negative maxAge, a zero evaluation time, a future observation, or a
// status older than maxAge all fail closed. Equality at the maximum allowed age
// remains acceptable.
func (s CheckpointStatus) ReadyForSubmissionWithin(submission CheckpointSubmission, now time.Time, maxAge time.Duration) bool {
	if maxAge <= 0 || now.IsZero() || !s.ReadyForSubmission(submission) {
		return false
	}
	observedAt := s.ObservedAt.UTC()
	evaluatedAt := now.UTC()
	if observedAt.After(evaluatedAt) {
		return false
	}
	return evaluatedAt.Sub(observedAt) <= maxAge
}
