package immich

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJobStatusStruct(t *testing.T) {
	resp := `{"thumbnailGeneration":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"metadataExtraction":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"videoConversion":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"objectTagging":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"recognizeFaces":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"clipEncoding":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"backgroundTask":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"storageTemplateMigration":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"search":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}},"sidecar":{"jobCounts":{"active":0,"completed":0,"failed":0,"delayed":0,"waiting":0,"paused":0},"queueStatus":{"isActive":false,"isPaused":false}}}`

	jobs := JobStatuses{}
	err := json.NewDecoder(strings.NewReader(resp)).Decode(&jobs)
	if err != nil {
		t.Error(err)
	}
}
