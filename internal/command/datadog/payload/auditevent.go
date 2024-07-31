package payload

import (
	"encoding/json"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
)

type AuditEvent struct {
	Data *string

	Common
}

func (auditEvent AuditEvent) Enrich(header http.Header, evt *datadogsender.Event, logger *zap.SugaredLogger) {
	auditEvent.Common.Enrich(header, evt, logger)

	if data := auditEvent.Data; data != nil {
		var auditEventData auditEventData

		if err := json.Unmarshal([]byte(*data), &auditEventData); err != nil {
			logger.Warnf("failed to unmarshal audit event's data: %v", err)

			return
		} else {
			auditEventData.Enrich(header, evt, logger)
		}
	}
}
