// Code generated by "stringer -type JobID -trimprefix Job"; DO NOT EDIT.

package immich

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[JobThumbnailGeneration-0]
	_ = x[JobMetadataExtraction-1]
	_ = x[JobVideoConversion-2]
	_ = x[JobObjectTagging-3]
	_ = x[JobRecognizeFaces-4]
	_ = x[JobClipEncoding-5]
	_ = x[JobBackgroundTask-6]
	_ = x[JobStorageTemplateMigration-7]
	_ = x[JobSearch-8]
	_ = x[JobSidecar-9]
}

const _JobID_name = "ThumbnailGenerationMetadataExtractionVideoConversionObjectTaggingRecognizeFacesClipEncodingBackgroundTaskStorageTemplateMigrationSearchSidecar"

var _JobID_index = [...]uint8{0, 19, 37, 52, 65, 79, 91, 105, 129, 135, 142}

func (i JobID) String() string {
	if i < 0 || i >= JobID(len(_JobID_index)-1) {
		return "JobID(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _JobID_name[_JobID_index[i]:_JobID_index[i+1]]
}