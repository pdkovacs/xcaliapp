package vcblobstore

import "time"

type BlobInfo struct {
	Key        string
	Content    []byte
	ModifiedBy string
}

type BlobVersion struct {
	VersionID      string
	ModifiedBy     string
	ModifiedAt     time.Time
	Size           int64
	IsLatest       bool
	IsDeleteMarker bool
}
