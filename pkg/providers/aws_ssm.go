package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/danielr18/teller/pkg/core"
	"github.com/danielr18/teller/pkg/logging"
)

type AWSSSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}
type AWSSSM struct {
	client AWSSSMClient
	logger logging.Logger
}

const awsssmName = "aws_ssm"

// nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "AWS SSM (aka paramstore)",
		Name:           awsssmName,
		Authentication: "Your standard `AWS_DEFAULT_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` need to be populated in your environment. `AWS_ENDPOINT` is used to allow usage of localstack",
		ConfigTemplate: `
  # configure only from environment
  aws_ssm:
    env:
      FOO_BAR:
        path: /prod/foobar
        decrypt: true
		`,
		Ops: core.OpMatrix{Get: true},
	}
	RegisterProvider(metaInfo, NewAWSSSM)
}

func NewAWSSSM(logger logging.Logger) (core.Provider, error) {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		awsEndpointOverride := os.Getenv("AWS_ENDPOINT")
		if awsEndpointOverride != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpointOverride,
				SigningRegion: region,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolver(customResolver))
	if err != nil {
		return nil, err
	}

	client := ssm.NewFromConfig(cfg)

	return &AWSSSM{client: client, logger: logger}, nil
}

func (a *AWSSSM) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", awsssmName)
}
func (a *AWSSSM) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", awsssmName)
}

func (a *AWSSSM) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("does not support full env sync (path: %s)", kp.Path)
}

func (a *AWSSSM) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", awsssmName)
}

func (a *AWSSSM) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", awsssmName)
}

func (a *AWSSSM) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		a.logger.WithField("path", p.Path).Debug("secret is empty")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*secret)
	return &ent, nil
}

func (a *AWSSSM) getSecret(kp core.KeyPath) (*string, error) {
	a.logger.WithField("path", kp.Path).Debug("get entry")
	res, err := a.client.GetParameter(context.TODO(), &ssm.GetParameterInput{Name: &kp.Path, WithDecryption: kp.Decrypt})
	if err != nil {
		return nil, err
	}

	if res == nil || res.Parameter.Value == nil {
		return nil, nil
	}

	return res.Parameter.Value, nil
}
