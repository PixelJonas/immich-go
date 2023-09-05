package immich

import (
	"context"
	"time"
)

func (ic *ImmichClient) StartAndWaitJob(ctx context.Context, job JobID, cmd JobCommand, force bool) error {
	_, err := ic.StartJob(ctx, job, cmd, force)
	if err != nil {
		return err
	}
	jobName := idToString(job)

	for {
		time.Sleep(2 * time.Second)
		j, err := ic.GetJobs(ctx)
		if err != nil {
			return err
		}
		if j[jobName].Active == 0 {
			break
		}
	}
	return nil
}

func (ic *ImmichClient) StartJob(ctx context.Context, job JobID, cmd JobCommand, force bool) (jobStatus, error) {

	resp := jobStatus{}
	req := struct {
		Command string `json:"command"`
		Force   bool   `json:"force"`
	}{
		Command: idToString(cmd),
		Force:   force,
	}
	err := ic.newServerCall(ctx, "JobCommand").
		do(put("/jobs/"+idToString(job), setAcceptJSON(), setJSONBody(req)),
			responseJSON(&resp))
	return resp, err
}

func (ic *ImmichClient) GetJobs(ctx context.Context) (JobStatuses, error) {
	statuses := JobStatuses{}
	err := ic.newServerCall(ctx, "GetJobs").
		do(get("/jobs", setAcceptJSON()), responseJSON(&statuses))
	return statuses, err
}

//go:generate stringer -type JobID -trimprefix Job
type JobID int

const (
	JobThumbnailGeneration JobID = iota
	JobMetadataExtraction
	JobVideoConversion
	JobObjectTagging
	JobRecognizeFaces
	JobClipEncoding
	JobBackgroundTask
	JobStorageTemplateMigration
	JobSearch
	JobSidecar
)

//go:generate stringer -type JobCommand -trimprefix Cmd
type JobCommand int

const (
	CmdStart JobCommand = iota
	CmdPause
	CmdResume
	CmdEmpty
)

func idToString[T interface{ String() string }](id T) string {
	s := id.String()
	return string(s[0]+'a'-'A') + s[1:]
}

type QueueStatus struct {
	IsActive bool `json:"isActive"`
	IsPaused bool `json:"isPaused"`
}
type JobCounts struct {
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Delayed   int `json:"delayed"`
	Waiting   int `json:"waiting"`
	Paused    int `json:"paused"`
}
type jobStatus struct {
	JobCounts   `json:"jobCounts"`
	QueueStatus `json:"queueStatus"`
}

type JobStatuses map[string]jobStatus
