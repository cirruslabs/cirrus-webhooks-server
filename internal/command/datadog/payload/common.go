package payload

import (
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type common struct {
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
}

func (common common) Enrich(header http.Header, evt *datadogsender.Event, logger *zap.SugaredLogger) {
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
}
