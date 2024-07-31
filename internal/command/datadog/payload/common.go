package payload

import (
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Common struct {
	Action    *string `json:"action"`
	Type      *string `json:"type"`
	Timestamp *int64  `json:"timestamp"`
	Actor     struct {
		ID *int64 `json:"id"`
	}
	Repository struct {
		ID    *int64  `json:"id"`
		Owner *string `json:"owner"`
		Name  *string `json:"name"`
	}
	Build struct {
		ID          *int64  `json:"id"`
		Status      *string `json:"status"`
		Branch      *string `json:"branch"`
		PullRequest *int64  `json:"pullRequest"`
		User        struct {
			Username *string `json:"username"`
		} `json:"user"`
	}
	Task struct {
		ID           *int64   `json:"id"`
		Name         *string  `json:"name"`
		Status       *string  `json:"status"`
		InstanceType *string  `json:"instanceType"`
		UniqueLabels []string `json:"uniqueLabels"`
	}
}

func (common Common) Enrich(header http.Header, evt *datadogsender.Event, logger *zap.SugaredLogger) {
	if value := common.Action; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("action:%s", *value))
	}

	if t := common.Type; t != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("type:%s", *t))
	}

	if rawTimestamp := header.Get("X-Cirrus-Timestamp"); rawTimestamp != "" {
		timestamp, err := strconv.ParseInt(rawTimestamp, 10, 64)
		if err != nil {
			logger.Warnf("failed to parse \"X-Cirrus-Timestamp\" timestamp value %q: %v",
				rawTimestamp, err)
		} else {
			evt.Timestamp = time.UnixMilli(timestamp)
		}
	}

	if value := common.Actor.ID; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("actor_id:%d", *value))
	}

	if value := common.Repository.ID; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("repository_id:%d", *value))
	}
	if value := common.Repository.Owner; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("repository_owner:%s", *value))
	}
	if value := common.Repository.Name; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("repository_name:%s", *value))
	}

	if value := common.Build.ID; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_id:%d", *value))
	}
	if value := common.Build.Status; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_status:%s", *value))
	}
	if value := common.Build.Branch; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_branch:%s", *value))
	}
	if value := common.Build.PullRequest; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("build_pull_request:%d", *value))
	}
	if value := common.Build.User.Username; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("initializer_username:%s", *value))
	}

	if value := common.Task.ID; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_id:%d", *value))
	}
	if value := common.Task.Name; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_name:%s", *value))
	}
	if value := common.Task.Status; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_status:%s", *value))
	}
	if value := common.Task.InstanceType; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_instance_type:%s", *value))
	}
	if value := common.Task.UniqueLabels; len(value) > 0 {
		evt.Tags = append(evt.Tags, fmt.Sprintf("task_unique_labels:%s", strings.Join(value, ",")))
	}
}
