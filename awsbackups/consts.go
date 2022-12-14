package awsbackups

const (
	//timing
	hrsInitBackup   int = 48
	hrsConfInit     int = 0
	hrsDownloadJira int = 3
	hrsDownloadConf int = 3
	//s3 strings
	statePath      string = "lambda-state"
	stateFileName  string = "state.json"
	confS3BackPath string = "confluence/"
	jiraS3BackPath string = "jira/"
	//ActionStrings
	actionInitJira  string = "init-backup"
	actionInitConf  string = "init-confluence-backup"
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
