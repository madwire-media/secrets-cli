package types

// FetchedSecret represents a secret that has been fetched from its remote
// repository
type FetchedSecret interface {
	Value() interface{}
	Version() interface{}
	Format() int
	IsMissingData() bool

	UploadNew(value interface{}) (interface{}, error)
}
