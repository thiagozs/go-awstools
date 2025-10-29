package awstools

type Options func(*AWSToolsParams) error

type AWSToolsParams struct {
	region       string
	accessKeyID  string
	secretKey    string
	sessionToken string
	bufferLimit  int // limit buffer channel for read file
	workersRLS   int // amount of worker read line Stream
	endpoint     string
	disableSSL   bool
}

func newAWSToolsParams(opts ...Options) (*AWSToolsParams, error) {
	awsToolsParams := &AWSToolsParams{}
	for _, opt := range opts {
		if err := opt(awsToolsParams); err != nil {
			return nil, err
		}
	}
	// set sensible defaults
	if awsToolsParams.bufferLimit == 0 {
		awsToolsParams.bufferLimit = 100
	}
	if awsToolsParams.workersRLS == 0 {
		awsToolsParams.workersRLS = 4
	}
	return awsToolsParams, nil
}

func WithEndpoint(endpoint string) Options {
	return func(p *AWSToolsParams) error {
		p.endpoint = endpoint
		return nil
	}
}

func WithRegion(region string) Options {
	return func(p *AWSToolsParams) error {
		p.region = region
		return nil
	}
}

func WithAccessKeyID(accessKeyID string) Options {
	return func(p *AWSToolsParams) error {
		p.accessKeyID = accessKeyID
		return nil
	}
}

func WithSecretKey(secretKey string) Options {
	return func(p *AWSToolsParams) error {
		p.secretKey = secretKey
		return nil
	}
}

func WithSessionToken(sessionToken string) Options {
	return func(p *AWSToolsParams) error {
		p.sessionToken = sessionToken
		return nil
	}
}

func WithBufferLimit(bufferLimit int) Options {
	return func(p *AWSToolsParams) error {
		p.bufferLimit = bufferLimit
		return nil
	}
}

func WithAmountWorkersRLS(workersRLS int) Options {
	return func(p *AWSToolsParams) error {
		p.workersRLS = workersRLS
		return nil
	}
}

func WithDisableSSL(disable bool) Options {
	return func(p *AWSToolsParams) error {
		p.disableSSL = disable
		return nil
	}
}

// getters -----

func (p *AWSToolsParams) Region() string {
	return p.region
}

func (p *AWSToolsParams) AccessKeyID() string {
	return p.accessKeyID
}

func (p *AWSToolsParams) SecretKey() string {
	return p.secretKey
}

func (p *AWSToolsParams) SessionToken() string {
	return p.sessionToken
}

func (p *AWSToolsParams) BufferLimit() int {
	return p.bufferLimit
}

func (p *AWSToolsParams) AmountWorkersRLS() int {
	return p.workersRLS
}

func (p *AWSToolsParams) Endpoint() string {
	return p.endpoint
}

func (p *AWSToolsParams) DisableSSL() bool {
	return p.disableSSL
}

// setters -----

func (p *AWSToolsParams) SetRegion(region string) {
	p.region = region
}

func (p *AWSToolsParams) SetAccessKeyID(accessKeyID string) {
	p.accessKeyID = accessKeyID
}

func (p *AWSToolsParams) SetSecretKey(secretKey string) {
	p.secretKey = secretKey
}

func (p *AWSToolsParams) SetSessionToken(sessionToken string) {
	p.sessionToken = sessionToken
}

func (p *AWSToolsParams) SetBufferLimit(bufferLimit int) {
	p.bufferLimit = bufferLimit
}

func (p *AWSToolsParams) SetAmountWorkersRLS(workersRLS int) {
	p.workersRLS = workersRLS
}

func (p *AWSToolsParams) SetEndpoint(endpoint string) {
	p.endpoint = endpoint
}
