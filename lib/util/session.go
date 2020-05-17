package util

import "github.com/aws/aws-sdk-go/aws/session"

func CreateSession(profile string) *session.Session {
	if profile != "" {
		return session.Must(
			session.NewSessionWithOptions(
				session.Options{
					Profile: profile,
				},
			),
		)
	} else {
		return session.Must(session.NewSession())
	}
}
