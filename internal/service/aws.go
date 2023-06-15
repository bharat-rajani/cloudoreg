package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	// TODO: Change SID?
	CloudoregPolicy = `{
	  "Version": "2012-10-17",
	  "Statement": [
		{
		  "Sid": "CloudigradePolicy",
		  "Effect": "Allow",	
		  "Action": [
			"sts:GetCallerIdentity"
		  ],
		  "Resource": "*"
		}
	  ]
	}`
)

type AWSService interface {
	AssumeRole(awsARN arn.ARN) (*credentials.Credentials, error)
	GetCallerIdentity(sess ...*session.Session) (*sts.GetCallerIdentityOutput, error)
	CreateCustomerSession(creds *credentials.Credentials) (*session.Session, error)
}

type AWSServiceImpl struct {
	AppSession *session.Session
}

func (as *AWSServiceImpl) AssumeRole(awsARN arn.ARN) (*credentials.Credentials, error) {

	svc := sts.New(as.AppSession)
	input := &sts.AssumeRoleInput{
		Policy:          aws.String(CloudoregPolicy),
		RoleArn:         aws.String(awsARN.String()),
		RoleSessionName: aws.String(fmt.Sprintf("cloudoreg-%s", awsARN.AccountID)),
	}

	result, err := svc.AssumeRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeMalformedPolicyDocumentException:
				return nil, fmt.Errorf("AWS service exception %s, error: %w", sts.ErrCodeMalformedPolicyDocumentException, aerr)
			case sts.ErrCodePackedPolicyTooLargeException:
				return nil, fmt.Errorf("AWS service exception %s, error: %w", sts.ErrCodePackedPolicyTooLargeException, aerr)
			case sts.ErrCodeRegionDisabledException:
				return nil, fmt.Errorf("AWS service exception %s, error: %w", sts.ErrCodeRegionDisabledException, aerr)
			case sts.ErrCodeExpiredTokenException:
				return nil, fmt.Errorf("AWS service exception %s, error: %w", sts.ErrCodeExpiredTokenException, aerr)
			default:
				return nil, fmt.Errorf("AWS service error: %w", aerr)
			}
		}
		return nil, fmt.Errorf("AWS service error: %w", err)
	}

	creds := credentials.NewStaticCredentials(*result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken)
	return creds, nil
}

func (as *AWSServiceImpl) GetCallerIdentity(s ...*session.Session) (*sts.GetCallerIdentityOutput, error) {
	sess := as.AppSession
	if s != nil {
		sess = s[0]
	}
	svc := sts.New(sess)
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		return nil, fmt.Errorf("AWS service error: %w", err)
	}

	return result, nil
}

func (as *AWSServiceImpl) CreateCustomerSession(creds *credentials.Credentials) (*session.Session, error) {
	newSession, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		return nil, err
	}
	return newSession, nil
}
