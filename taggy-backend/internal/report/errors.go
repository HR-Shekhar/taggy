package report

import "errors"

var (
	ErrReportNotFound   = errors.New("report not found")
	ErrDuplicateOpen    = errors.New("an open report already exists for this target")
	ErrInvalidTarget    = errors.New("invalid report target")
	ErrInvalidReason    = errors.New("report reason is invalid")
	ErrCannotReportSelf = errors.New("cannot report yourself")
)
