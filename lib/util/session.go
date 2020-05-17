package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func CreateSession(profile string, region string) *session.Session {
	if profile == "" && region == "" {
		// default
		return session.Must(session.NewSession())
	} else if profile == "" && region != "" {
		// with region
		return session.Must(
			session.NewSessionWithOptions(
				session.Options{
					Config: aws.Config{
						Region: aws.String(region),
					},
				},
			),
		)
	} else if profile != "" && region == "" {
		// with profile
		return session.Must(
			session.NewSessionWithOptions(
				session.Options{
					Profile: profile,
				},
			),
		)
	} else {
		// with region and profile
		return session.Must(
			session.NewSessionWithOptions(
				session.Options{
					Config: aws.Config{
						Region: aws.String(region),
					},
					Profile: profile,
				},
			),
		)
	}
}
