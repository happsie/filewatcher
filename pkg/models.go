package pkg

type ModificationType int32

const (
	Modified ModificationType = 0
	Deleted  ModificationType = 1
	Created  ModificationType = 2
)

type WatchEvent struct {
	ModificationType ModificationType
	Name             string
}

