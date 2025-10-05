package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ground truth: docs/api/json_rpc_methods.md §Event Subscription Methods
// Reuse client methods: SubscribeEvents, UnsubscribeEvents, GetSubscriptionStats

func TestSubscribeUnsubscribeEvents(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	// Subscribe to a couple of topics
	subResp, err := fixture.client.SubscribeEvents([]string{"camera.connected", "recording.start"})
	require.NoError(t, err)
	require.Nil(t, subResp.Error)

	result := subResp.Result.(map[string]interface{})
	assert.Contains(t, result, "subscribed")
	assert.Contains(t, result, "topics")

	// Get stats
	statsResp, err := fixture.client.GetSubscriptionStats()
	require.NoError(t, err)
	require.Nil(t, statsResp.Error)

	// Unsubscribe one topic
	unsubResp, err := fixture.client.UnsubscribeEvents([]string{"camera.connected"})
	require.NoError(t, err)
	require.Nil(t, unsubResp.Error)
}

func TestEventDeliveryEnvelope_Basic(t *testing.T) {
	// Minimal check: after subscribing, server may deliver notifications.
	// We validate only that server accepts subscription and test harness remains stable.
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	subResp, err := fixture.client.SubscribeEvents([]string{"system.startup"})
	require.NoError(t, err)
	require.Nil(t, subResp.Error)
}
