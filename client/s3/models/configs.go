package models

type S3Configs struct {

	Region string
	Key string
	DownloadLocation string
	Bucket string
	EncryptionAlgorithm string
	PartSize int64
}