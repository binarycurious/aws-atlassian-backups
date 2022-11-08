package awsbackups

// ExecEvent ...
// AWS Event passed by lambda exectution
type ExecEvent struct {
	Name string `json:"name"`
}

type lambdaState struct {
	LastAction     string `json:"lastAction"`
	LastExecution  string `json:"lastExecution"`
	LastJiraTaskID string `json:"lastJiraTaskID"`
	LastResult     string `json:"lastResult"`
	ErrData        string `json:"errData"`
}

type progressData struct {
	//Confluence progress fields
	FileName                   string `json:"fileName"`
	Size                       int    `json:"size"`
	CurrentStatus              string `json:"currentStatus"`
	AlternativePercentage      string `json:"alternativePercentage"`
	ConcurrentBackupInProgress bool   `json:"concurrentBackupInProgress"`
	Time                       int64  `json:"time"`
	IsOutdated                 bool   `json:"isOutdated"`
	//Jira progress fields
	Status      string `json:"status"`      //: "Success",
	Description string `json:"description"` //: "Cloud Export task",
	Message     string `json:"message"`     //: "Completed export",
	Result      string `json:"result"`      //: "export/download/?fileId=3265d317-f699-4bca-be39-3ead8b4453f3",
	Progress    int    `json:"progress"`    //: 100,
	ExportType  string `json:"exportType"`  //: "CLOUD"
}

type envConfig struct {
	email      string
	apiToken   string
	hostname   string
	bucketName string
}
