# jira-backup-go

Simple Go Module for running Jira and Confluence Backups in AWS via Lambda if small enough or if you require a longer running download (> 15m) you can use AWS ECS Fargate Tasks

The process does not require intermediate storage volumes and will stream the download straight to S3

The consts module file contains various values that relate to how the state file and backups will be managed in S3 and will need some modification depending on your setup.

You will need to add a .env file to your runtime environment that contains the variables for your Atalassian API or you will need to set these environment variables another way.

"API_EMAIL"
"API_TOKEN"
"API_HOSTNAME"

