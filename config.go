package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jahkeup/updater53/pkg/whatip"
)

// Config holds the run configuration for the update.
type Config struct {
	Records []string
	IPer    whatip.IPer
	Session *session.Session
	Commit  bool
}
