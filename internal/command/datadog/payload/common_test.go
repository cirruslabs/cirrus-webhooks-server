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

func TestEnrichBuild(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "build.json"))
	require.NoError(t, err)

	evt := &datadogsender.Event{}

	payload := payload.Common{}
	require.NoError(t, json.Unmarshal(body, &payload))
	payload.Enrich(http.Header{
		"X-Cirrus-Timestamp": []string{strconv.FormatInt(time.Now().UnixMilli(), 10)},
	}, evt, zap.S())
	require.WithinDuration(t, time.Now(), evt.Timestamp, time.Second)
	require.Equal(t, []string{
		"action:updated",
		"repository_id:5129885287448576",
		"repository_owner:edigaryev",
		"repository_name:awesome-system-calls",
		"build_id:5082236150611968",
		"build_status:EXECUTING",
		"build_branch:main",
		"initializer_username:edigaryev",
	}, evt.Tags)
}

func TestEnrichTask(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "task.json"))
	require.NoError(t, err)

	evt := &datadogsender.Event{}

	payload := payload.Common{}
	require.NoError(t, json.Unmarshal(body, &payload))
	payload.Enrich(http.Header{
		"X-Cirrus-Timestamp": []string{strconv.FormatInt(time.Now().UnixMilli(), 10)},
	}, evt, zap.S())
	require.WithinDuration(t, time.Now(), evt.Timestamp, time.Second)
	require.Equal(t, []string{
		"action:created",
		"repository_id:5129885287448576",
		"repository_owner:edigaryev",
		"repository_name:awesome-system-calls",
		"build_id:5082236150611968",
		"build_status:EXECUTING",
		"build_branch:main",
		"initializer_username:edigaryev",
		"task_id:6017965227769856",
		"task_name:Lint (cargo fmt)",
		"task_status:EXECUTING",
		"task_instance_type:CommunityContainer",
	}, evt.Tags)
}
