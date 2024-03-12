package backend

type BackendConfig struct {
	Enabled    bool             `json:"enabled" yaml:"enabled"`
	Store      StoreConfig      `json:"store" yaml:"store"`
	Encryption EncryptionConfig `json:"encryption" yaml:"encryption"`
}

type StoreConfig struct {
	AwsS3 *AwsS3BackendConfig `json:"awsS3,omitempty" yaml:"awsS3,omitempty"`
}

type EncryptionConfig struct {
	AwsKms *AwsKmsBackendConfig `json:"awsKms,omitempty" yaml:"awsKms,omitempty"`
}

type AwsKmsBackendConfig struct {
	KmsKey string `json:"kmsKey" yaml:"kmsKey"`
}

type AwsS3BackendConfig struct {
	Bucket string `json:"bucket" yaml:"bucket"`
}
