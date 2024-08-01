package payload_test

import (
	"encoding/json"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog/payload"
	"github.com/cirruslabs/cirrus-webhooks-server/internal/datadogsender"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestEnrichAuditEvent(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "audit_event.json"))
	require.NoError(t, err)

	evt := &datadogsender.Event{}

	payload := payload.AuditEvent{}
	require.NoError(t, json.Unmarshal(body, &payload))
	payload.Enrich(http.Header{
		"X-Cirrus-Timestamp": []string{strconv.FormatInt(time.Now().UnixMilli(), 10)},
	}, evt, zap.S())
	require.WithinDuration(t, time.Now(), evt.Timestamp, time.Second)
	require.Equal(t, []string{
		"action:created",
		"type:graphql.mutation",
		"data.mutationName:GenerateNewScopedAccessToken",
		"actor_username:edigaryev",
		"actor_location_ip:1.2.3.4",
	}, evt.Tags)
}
