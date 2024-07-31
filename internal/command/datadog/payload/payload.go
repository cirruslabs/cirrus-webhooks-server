package payload

import (
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"go.uber.org/zap"
	"net/http"
)

type Payload interface {
	Enrich(http.Header, *datadogsender.Event, *zap.SugaredLogger)
}
