package account

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusArchived:
		return true
	default:
		return false
	}
}
