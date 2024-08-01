package payload

import (
	"encoding/json"
	"fmt"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
)

type AuditEvent struct {
	Data *string `json:"data"`

	Actor struct {
		Username *string `json:"username"`
	} `json:"actor"`

	ActorLocationIP *string `json:"actorLocationIp"`

	common
}

func (auditEvent AuditEvent) Enrich(header http.Header, evt *datadogsender.Event, logger *zap.SugaredLogger) {
	auditEvent.common.Enrich(header, evt, logger)

	if data := auditEvent.Data; data != nil {
		var auditEventData auditEventData

		if err := json.Unmarshal([]byte(*data), &auditEventData); err != nil {
			logger.Warnf("failed to unmarshal audit event's data: %v", err)

			return
		} else {
			auditEventData.Enrich(header, evt, logger)
		}
	}

	actorUsername := "api"
	if value := auditEvent.Actor.Username; value != nil {
		actorUsername = *value
	}
	evt.Tags = append(evt.Tags, fmt.Sprintf("actor_username:%s", actorUsername))

	if value := auditEvent.ActorLocationIP; value != nil {
		evt.Tags = append(evt.Tags, fmt.Sprintf("actor_location_ip:%s", *value))
	}
}
