package bucketS3

const s3ServerRegion = "us-east-1"

type Config struct {
	S3StorageAddress string
}

type RunConfig struct {
	ServiceStackName string
	BucketName       string
	XAmzAcl          string
	AccessKeyId      string
	SecretAccessKey  string
}

type Handler struct {
	config Config
}

func New(
	config Config,
) *Handler {
	return &Handler{
		config: config,
	}
}
