# jira-backup-go

Simple Go Module for running Jira and Confluence Backups in AWS via Lambda if small enough or if you require a longer running download (> 15m) you can use AWS ECS Fargate Tasks

The process does not require intermediate storage volumes and will stream the download straight to S3

The consts module file contains various values that relate to how the state file and backups will be managed in S3.

You will need to add a .env file to your runtime environment that contains the variables for your Atalassian API and S3 bucket name or you will need to set these environment variables another way.

"API_EMAIL"
"API_TOKEN"
"API_HOSTNAME"
"AWS_S3_BUCKETNAME"

This module is based on information provided by Atlassian which is not in LTS, this module may need to be updated if the Atlassian backup process changes. See notes at : https://confluence.atlassian.com/jirakb/automate-backups-for-jira-cloud-779160659.html

