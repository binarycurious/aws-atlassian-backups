# jira-backup-go

Simple Go Module for running Jira and Confluence Backups in AWS via Lambda if small enough or if you require a longer running download (> 15m) you can use AWS ECS Fargate Tasks

The process does not require intermediate storage volumes and will stream the download straight to S3