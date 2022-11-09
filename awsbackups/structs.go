package awsbackups

import (
	"fmt"
	"time"
)

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

func (s *lambdaState) goodToGo(action string) bool {

	t, err := time.Parse(time.RFC3339, s.LastExecution)
	if err != nil {
		fmt.Println("Could not parse last exec. time")
		fmt.Println(err)
		t = time.Now().Add(time.Duration(time.Hour * 72 / -1))
	}

	switch action {
	case actionInit:
		return t.Add(time.Hour * time.Duration(hrsInitBackup)).Before(time.Now())
	case actionSaveJira:
		return t.Add(time.Hour * time.Duration(hrsDownloadJira)).Before(time.Now())
	case actionSaveConf:
		return t.Add(time.Hour * time.Duration(hrsDownloadConf)).Before(time.Now())
	}
	return false
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
