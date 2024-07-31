package payload

import (
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
)

type auditEventData struct {
	MutationName *string `json:"mutationName"`
	BuildID      *string `json:"buildId"`
	TaskID       *string `json:"taskId"`
}

func (auditEventData auditEventData) Enrich(header http.Header, evt *datadogsender.Event, logger *zap.SugaredLogger) {
	if mutationName := auditEventData.MutationName; mutationName != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("data.mutationName:%s", *mutationName))
	}

	if buildID := auditEventData.BuildID; buildID != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("data.buildId:%s", *buildID))
	}

	if taskID := auditEventData.TaskID; taskID != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("data.taskId:%s", *taskID))
	}
}
