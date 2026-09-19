package protection

import "errors"

var (
	errDatasetIDEmpty = errors.New("dataset ID must not be empty")
	errEvaluationTimeZero = errors.New("evaluation time must not be zero")
	errEvidenceKindEmpty = errors.New("evidence kind must not be empty")
	errObservedTimeZero = errors.New("observed time must not be zero")
	errPolicyIDEmpty = errors.New("policy ID must not be empty")
	errRecoveryPointIDEmpty = errors.New("recovery point ID must not be empty")
	errRepositoryIDEmpty = errors.New("repository ID must not be empty")
	errRestoreFailureCategoryMissing = errors.New("failing restore test must have a failure category")
	errRestoreTestCompletionTimeZero = errors.New("restore test completion time must not be zero")
	errRestoreVerificationMaxAgeNegative = errors.New("restore verification max age must not be negative")
	errRestoreVerificationObservedAfterEvaluation = errors.New("restore verification observation time is after evaluation time")
)
