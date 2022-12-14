package awsbackups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/binarycurious/go-string-helpers/stringhelpers"
	"github.com/joho/godotenv"
)

var config envConfig

func init() {
	err := godotenv.Load("/etc/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	enVars := os.Environ()

	for itm := range enVars {
		if os.Getenv("DEBUG") == "true" {
			fmt.Printf(strings.Split(enVars[itm], "=")[0])
			fmt.Println(" = " + strings.Split(enVars[itm], "=")[1])
		}
	}

	bucketName := os.Getenv("AWS_S3_BUCKETNAME")
	s3Region := *stringhelpers.CoalesceWhitespace(os.Getenv("AWS_S3_REGION"), endpoints.ApSoutheast2RegionID)

	if bucketName == "" {
		log.Fatal("No S3 bucket name set for backups (AWS_S3_BUCKETNAME)")
	}

	config = envConfig{
		email:      os.Getenv("API_EMAIL"),
		apiToken:   os.Getenv("API_TOKEN"),
		hostname:   os.Getenv("API_HOSTNAME"),
		bucketName: bucketName,
		region:     s3Region,
		statePath:  *stringhelpers.CoalesceWhitespace(os.Getenv("AWS_S3_STATE_PATH"), statePath),
		jiraS3Path: *stringhelpers.CoalesceWhitespace(os.Getenv("AWS_S3_JIRA_PATH"), jiraS3BackPath),
		confS3Path: *stringhelpers.CoalesceWhitespace(os.Getenv("AWS_S3_CONFLUENCE_PATH"), confS3BackPath),
	}

}

// HandleRequest ...
// AWS Lambda event / request handler
func HandleRequest(ctx context.Context, event ExecEvent) (string, error) {
	fmt.Printf("Received event: %s\n", event.Name)
	fmt.Printf("Running backup with account: %s\v", config.email)

	var resp string
	var err error
	s := pullState()

	fmt.Printf("Executing with state: %#v", s)

	switch s.LastAction {
	case actionInitJira:
		if s.LastResult != stateOK {
			resp, err = initJira(&s)
		} else {
			resp, err = initConf(&s)
			s.LastAction = actionInitConf
		}

	case actionInitConf:
		if s.LastResult != stateOK {
			resp, err = initConf(&s)
		} else {
			resp, err = saveJiraBackup(&s)
			s.LastAction = actionSaveJira
		}

	case actionSaveJira:
		if s.LastResult != stateOK {
			resp, err = saveJiraBackup(&s)
		} else {
			resp, err = saveConfBackup(&s)
			s.LastAction = actionSaveConf
		}

	case actionSaveConf:
		if s.LastResult != stateOK {
			resp, err = saveConfBackup(&s)
		} else {
			resp, err = initJira(&s)
			s.LastAction = actionInitJira
		}
	}

	if err == nil {
		s.LastResult = stateOK
		s.ErrData = ""
	} else {
		s.LastResult = stateFailed
		s.ErrData += fmt.Sprintln(err)
	}

	if resp == stateWait {
		fmt.Println("Exiting execution on wait condition")
		return stateWait, nil
	}

	s.LastExecution = time.Now().Format(time.RFC3339)

	saveState(&s)

	return resp, err
}

func saveState(s *lambdaState) {
	fmt.Printf("Saving state file : %#v\n", s)

	j, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}

	uploadS3Stream(config.statePath, stateFileName, bytes.NewReader(j), s, actionPushState)
}

func createAPIRequest(path string, body string, s *lambdaState, actionInProc string, method string) *http.Request {
	r, err := http.NewRequest(method,
		fmt.Sprintf("https://%s%s", config.hostname, path),
		strings.NewReader(body),
	)
	if err != nil {
		failProc(s, actionInProc, err)
	}

	r.SetBasicAuth(config.email, config.apiToken)
	r.Header.Add("content-type", "application/json")

	return r
}

func initJira(s *lambdaState) (string, error) {
	if !s.goodToGo(actionInitJira) {
		fmt.Println("Waiting for execution delay, last execution " + s.LastExecution)
		return stateWait, nil
	}

	fmt.Println("Exectuing backup init...")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		jiraBakReq := createAPIRequest(
			urlJiraInit,
			fmt.Sprintf("{\"cbAttachments\":\"%s\", \"exportToCloud\":\"%s\"}",
				os.Getenv("INCLUDE_ATTACHMENTS"),
				os.Getenv("EXPORT_TO_CLOUD"),
			),
			s,
			actionInitJira,
			http.MethodPost,
		)

		resp, err := http.DefaultClient.Do(jiraBakReq)
		if err != nil {
			failProc(s, actionInitJira, err)
		}

		resBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			failProc(s, actionInitJira, err)
		}

		if strings.Contains(string(resBody), "\"error\"") {
			failProc(s, actionInitJira, fmt.Errorf("Failed to initialize the backup in Jira %s", string(resBody)))
		}

		fmt.Println("Successfully initialized Jira backup process...")

		wg.Done()
	}()

	wg.Wait()

	return "success", nil
}

func initConf(s *lambdaState) (string, error) {
	if !s.goodToGo(actionInitConf) {
		fmt.Println("Waiting for execution delay, last execution " + s.LastExecution)
		return stateWait, nil
	}

	fmt.Println("Exectuing backup init...")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		confBakReq := createAPIRequest(
			urlConfInit,
			fmt.Sprintf("{\"cbAttachments\":\"%s\"}",
				os.Getenv("INCLUDE_ATTACHMENTS"),
			),
			s,
			actionInitConf,
			http.MethodPost,
		)
		confBakReq.Header.Add("X-Atlassian-Token", "no-check")
		confBakReq.Header.Add("X-Requested-With", "XMLHttpRequest")

		cr, err := http.DefaultClient.Do(confBakReq)
		if err != nil {
			failProc(s, actionInitConf, err)
		}

		confResp, err := ioutil.ReadAll(cr.Body)
		if err != nil {
			failProc(s, actionInitConf, err)
		}

		if strings.Contains(string(confResp), "backup") {
			failProc(s, actionInitConf, fmt.Errorf("Failed to initialize the backup in Confluence %s", string(confResp)))
		}

		fmt.Println("Successfully initialized Confluence backup process...")
		wg.Done()
	}()

	wg.Wait()

	return "success", nil
}

func saveJiraBackup(s *lambdaState) (string, error) {
	if !s.goodToGo(actionSaveJira) {
		fmt.Println("Waiting for execution delay, last execution " + s.LastExecution)
		return stateWait, nil
	}

	fmt.Println("Exectuing Jira backup save...")

	req := createAPIRequest(urlJiraTask, "", s, actionSaveConf, http.MethodGet)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		failProc(s, actionSaveJira, err)
	}
	if resp.StatusCode > 299 {
		failProc(s, actionSaveJira, fmt.Errorf("failed to get progress data (%s)", resp.Status))
	}

	taskID, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		failProc(s, actionSaveJira, err)
	}

	//Get the jira progress data
	pReq := createAPIRequest(urlJiraProg+string(taskID), "", s, actionSaveJira, http.MethodGet)

	r, err := http.DefaultClient.Do(pReq)
	if err != nil {
		failProc(s, actionSaveJira, err)
	}
	if r.StatusCode > 299 {
		failProc(s, actionSaveJira, fmt.Errorf("failed to get progress data (%s)", r.Status))
	}

	rawData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failProc(s, actionSaveJira, err)
	}
	var progData progressData
	err = json.Unmarshal(rawData, &progData)
	if err != nil {
		failProc(s, actionSaveJira, err)
	}

	if progData.Progress != 100 {
		failProc(s, actionSaveJira, fmt.Errorf("jira download not ready"))
	}

	dlReq := createAPIRequest(urlJiraDL+progData.Result, "", s, actionSaveJira, http.MethodGet)

	fmt.Printf("Attempting sync to S3 of jira backup : %s\n", urlJiraDL+progData.Result)

	dl, err := http.DefaultClient.Do(dlReq)
	if err != nil {
		failProc(s, actionSaveConf, err)
	}
	if dl.StatusCode > 299 {
		failProc(s, actionSaveConf, fmt.Errorf("failed to initiate download (%s)", dl.Status))
	}

	fileKey := fmt.Sprintf(fmtJiraFile, time.Now().Format(time.RFC3339))
	uploadS3Stream(config.jiraS3Path, fileKey, dl.Body, s, actionSaveConf)

	fmt.Printf("Completed jira backup file sync to S3 : s3://%s/%s/%s\n", config.bucketName, config.jiraS3Path, fileKey)
	return "success", nil
}

func saveConfBackup(s *lambdaState) (string, error) {
	if !s.goodToGo(actionSaveConf) {
		fmt.Println("Waiting for execution delay, last execution " + s.LastExecution)
		return stateWait, nil
	}

	fmt.Println("Exectuing Confluence backup save...")

	req := createAPIRequest(urlConfProg, "", s, actionSaveConf, http.MethodGet)

	cr, err := http.DefaultClient.Do(req)
	if err != nil {
		failProc(s, actionSaveConf, err)
	}
	if cr.StatusCode > 299 {
		failProc(s, actionSaveConf, fmt.Errorf("failed to get progress data (%s)", cr.Status))
	}

	rawData, err := ioutil.ReadAll(cr.Body)
	if err != nil {
		failProc(s, actionSaveConf, err)
	}
	var progData progressData
	err = json.Unmarshal(rawData, &progData)
	if err != nil {
		failProc(s, actionSaveConf, err)
	}

	if progData.AlternativePercentage != "100%" {
		failProc(s, actionSaveConf, fmt.Errorf("confluence download not ready"))
	}

	fmt.Printf("Attempting sync to S3 of confluence backup : %s\n", urlConfDL+progData.FileName)

	r2 := createAPIRequest(urlConfDL+progData.FileName, "", s, actionSaveConf, http.MethodGet)
	dl, err := http.DefaultClient.Do(r2)
	if err != nil {
		failProc(s, actionSaveConf, err)
	}
	if dl.StatusCode > 299 {
		failProc(s, actionSaveConf, fmt.Errorf("failed to initiate download (%s)", dl.Status))
	}

	fileKey := fmt.Sprintf(fmtConfFile, time.Now().Format(time.RFC3339))
	uploadS3Stream(config.confS3Path, fileKey, dl.Body, s, actionSaveConf)

	fmt.Printf("Completed confluence backup file sync to S3 : s3://%s/%s/%s\n", config.bucketName, config.confS3Path, fileKey)

	return "success", nil
}

func failProc(s *lambdaState, actionInProc string, err error) {

	if actionInProc == actionPushState {
		log.Fatal(err)
	}

	s.ErrData = fmt.Sprintln(err)
	s.LastResult = stateFailed
	s.LastExecution = time.Now().Local().String()
	s.LastAction = actionInProc

	saveState(s)

	log.Fatal(err)
}

func uploadS3Stream(path string, key string, stream io.Reader, s *lambdaState, actionInProc string) {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(config.region)}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.bucketName),
		Key:    aws.String(path + "/" + key),
		Body:   stream,
	})
	if err != nil {
		failProc(s, actionInProc, fmt.Errorf("failed to upload file %q from stream, %v", key, err))
	}

	fmt.Printf("file uploaded to, %s\n", result.Location)
}

func loadStateFromS3(path string, fileName string, s *lambdaState, actionInProc string) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(config.region)}))

	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})

	n, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(config.bucketName),
		Key:    aws.String(path + "/" + fileName),
	})

	json.Unmarshal(buf.Bytes(), s)

	if err != nil {
		log.Fatal(fmt.Sprintf("ERROR: failed to download file, %v", err))
	}
	fmt.Printf("file downloaded, %d bytes\n", n)
}

func pullState() lambdaState {
	var stateData lambdaState

	loadStateFromS3(config.statePath, stateFileName, &stateData, "")

	fmt.Println("Successfully Loaded " + stateFileName)

	return stateData
}
