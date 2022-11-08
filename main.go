package main

import (
	"github.com/binarycurious/aws-atlassian-backups/awsbackups"
)

func main() {
	awsbackups.HandleRequest(nil, awsbackups.ExecEvent{Name: "Exec"})
}
