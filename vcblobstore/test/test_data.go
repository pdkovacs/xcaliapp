package test

import (
	"vcblobstore"
	"crypto/rand"
)

func createTestBlob(key string, modifiedBy string) vcblobstore.BlobInfo {
	return vcblobstore.BlobInfo{
		Key:        key,
		Content:    randomBytes(4096),
		ModifiedBy: modifiedBy,
	}
}

func CloneBlob(blob vcblobstore.BlobInfo) vcblobstore.BlobInfo {
	contentClone := make([]byte, len(blob.Content))
	copy(contentClone, blob.Content)
	return vcblobstore.BlobInfo{
		Key:        blob.Key,
		Content:    contentClone,
		ModifiedBy: blob.ModifiedBy,
	}
}

func randomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

var TestData = []vcblobstore.BlobInfo{
	createTestBlob("metro-zazie", "ux"),
	createTestBlob("zazie-icon", "ux"),
}
