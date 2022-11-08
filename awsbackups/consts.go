package awsbackups

const (
	//timing
	hrsInitBackup   int = 48
	hrsDownloadJira int = 4
	hrsDownloadConf int = 4
	//s3 strings
	bucketName     string = "jiraandconfluencebackups"
	statePath      string = "lambda-state"
	stateFileName  string = "state.json"
	confS3BackPath string = "confluence/"
	jiraS3BackPath string = "jira/"
	//ActionStrings
	actionInit      string = "init-backup"
	actionSaveJira  string = "save-jira-backup"
	actionSaveConf  string = "save-confluence-backup"
	actionPushState string = "push-state"
	//Action state strings
	stateOK     string = "OK"
	stateFailed string = "failed"
	stateWait   string = "waiting"
	//atlassian api strings
	urlJiraInit string = "/rest/backup/1/export/runbackup"
	urlJiraTask string = "/rest/backup/1/export/lastTaskId"
	urlJiraProg string = "/rest/backup/1/export/getProgress?taskId="
	urlJiraDL   string = "/plugins/servlet/"
	urlConfInit string = "/wiki/rest/obm/1.0/runbackup"
	urlConfProg string = "/wiki/rest/obm/1.0/getprogress.json"
	urlConfDL   string = "/wiki/download/"
	//fmts
	fmtJiraFile string = "JIRA-backup-%s.zip"
	fmtConfFile string = "CONF-backup-%s.zip"
)
