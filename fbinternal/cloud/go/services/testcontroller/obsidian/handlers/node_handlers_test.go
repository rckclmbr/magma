/*
 * Copyright (c) Facebook, Inc. and its affiliates.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */

package handlers_test

import (
	"testing"
	"time"

	"magma/orc8r/cloud/go/clock"
	"magma/orc8r/cloud/go/obsidian"
	"magma/orc8r/cloud/go/obsidian/tests"
	"magma/orc8r/cloud/go/plugin"
	plugin2 "orc8r/fbinternal/cloud/go/plugin"
	"orc8r/fbinternal/cloud/go/services/testcontroller"
	"orc8r/fbinternal/cloud/go/services/testcontroller/obsidian/handlers"
	"orc8r/fbinternal/cloud/go/services/testcontroller/obsidian/models"
	"orc8r/fbinternal/cloud/go/services/testcontroller/storage"
	"orc8r/fbinternal/cloud/go/services/testcontroller/test_init"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func Test_ListCINodes(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	listNodes := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/nodes", obsidian.GET).HandlerFunc

	// Empty case
	tc := tests.Test{
		Method:         "GET",
		URL:            testURLRoot + "/nodes",
		Handler:        listNodes,
		ExpectedStatus: 200,
		ExpectedResult: tests.JSONMarshaler([]*models.CiNode{}),
	}
	tests.RunUnitTest(t, e, tc)

	// Happy path
	err := testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node1", VpnIP: "192.168.100.1"})
	assert.NoError(t, err)
	err = testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node2", VpnIP: "10.0.2.1"})
	assert.NoError(t, err)
	tc.ExpectedResult = tests.JSONMarshaler([]*models.CiNode{
		{
			Available:     swag.Bool(true),
			ID:            swag.String("node1"),
			LastLeaseTime: expectedDT(t, 0),
			VpnIP:         ipv4("192.168.100.1"),
		},
		{
			Available:     swag.Bool(true),
			ID:            swag.String("node2"),
			LastLeaseTime: expectedDT(t, 0),
			VpnIP:         ipv4("10.0.2.1"),
		},
	})
	tests.RunUnitTest(t, e, tc)
}

func Test_GetCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	getNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/nodes/:node_id", obsidian.GET).HandlerFunc

	// Empty case
	tc := tests.Test{
		Method:         "GET",
		URL:            testURLRoot + "/nodes/node1",
		Handler:        getNode,
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		ExpectedStatus: 404,
		ExpectedError:  "Not Found",
	}
	tests.RunUnitTest(t, e, tc)

	// Happy path
	err := testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node1", VpnIP: "192.168.100.1"})
	assert.NoError(t, err)
	tc = tests.Test{
		Method:         "GET",
		URL:            testURLRoot + "/nodes/node1",
		Handler:        getNode,
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		ExpectedStatus: 200,
		ExpectedResult: &models.CiNode{
			Available:     swag.Bool(true),
			ID:            swag.String("node1"),
			LastLeaseTime: expectedDT(t, 0),
			VpnIP:         ipv4("192.168.100.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)
}

func Test_CreateCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	createNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/nodes", obsidian.POST).HandlerFunc

	// Happy path
	tc := tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes",
		Handler:        createNode,
		ExpectedStatus: 201,
		Payload: &models.MutableCiNode{
			ID:    swag.String("node1"),
			VpnIP: ipv4("192.168.100.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)

	actual, err := testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected := map[string]*storage.CINode{
		"node1": {
			Id:            "node1",
			VpnIp:         "192.168.100.1",
			Available:     true,
			LastLeaseTime: timestampProto(t, 0),
		},
	}
	assert.Equal(t, expected, actual)
}

func Test_UpdateCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	updateNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/nodes/:node_id", obsidian.PUT).HandlerFunc

	// Happy path, create via PUT
	tc := tests.Test{
		Method:         "PUT",
		URL:            testURLRoot + "/nodes/node1",
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		Handler:        updateNode,
		ExpectedStatus: 204,
		Payload: &models.MutableCiNode{
			ID:    swag.String("node1"),
			VpnIP: ipv4("10.0.2.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)

	actual, err := testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected := map[string]*storage.CINode{
		"node1": {
			Id:            "node1",
			VpnIp:         "10.0.2.1",
			Available:     true,
			LastLeaseTime: timestampProto(t, 0),
		},
	}
	assert.Equal(t, expected, actual)

	// Happy path edit
	tc.Payload = &models.MutableCiNode{
		ID:    swag.String("node1"),
		VpnIP: ipv4("192.168.100.1"),
	}
	tests.RunUnitTest(t, e, tc)
	actual, err = testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected["node1"].VpnIp = "192.168.100.1"
	assert.Equal(t, expected, actual)

	// ID mismatch
	tc.Payload = &models.MutableCiNode{
		ID:    swag.String("node2"),
		VpnIP: ipv4("192.168.100.2"),
	}
	tc.ExpectedStatus = 400
	tc.ExpectedError = "payload ID does not match path param"
	tests.RunUnitTest(t, e, tc)
	actual, err = testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func Test_DeleteCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	deleteNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/nodes/:node_id", obsidian.DELETE).HandlerFunc

	// Empty case; don't error if DNE
	tc := tests.Test{
		Method:         "DELETE",
		URL:            testURLRoot + "/nodes/node1",
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		Handler:        deleteNode,
		ExpectedStatus: 204,
	}
	tests.RunUnitTest(t, e, tc)

	// Happy path
	err := testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node1", VpnIP: "10.0.2.1"})
	assert.NoError(t, err)
	tests.RunUnitTest(t, e, tc)
	actual, err := testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	assert.Empty(t, actual)
}

func Test_ReserveCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	frozenClock := 1000 * time.Hour
	clock.SetAndFreezeClock(t, time.Unix(0, 0).Add(frozenClock))
	defer clock.UnfreezeClock(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	reserveNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/reserve", obsidian.POST).HandlerFunc

	// Empty case
	tc := tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/reserve",
		Handler:        reserveNode,
		ExpectedStatus: 404,
		ExpectedError:  "Not Found",
	}
	tests.RunUnitTest(t, e, tc)

	// Happy path
	err := testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node1", VpnIP: "192.168.100.1"})
	assert.NoError(t, err)
	tc = tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/reserve",
		Handler:        reserveNode,
		ExpectedStatus: 200,
		ExpectedResult: &models.NodeLease{
			ID:      swag.String("node1"),
			LeaseID: swag.String("1"),
			VpnIP:   ipv4("192.168.100.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)
	actual, err := testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected := map[string]*storage.CINode{
		"node1": {
			Id:            "node1",
			VpnIp:         "192.168.100.1",
			Available:     false,
			LastLeaseTime: timestampProto(t, int64(frozenClock/time.Second)),
		},
	}
	assert.Equal(t, expected, actual)

	// Pool is empty
	tc = tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/reserve",
		Handler:        reserveNode,
		ExpectedStatus: 404,
		ExpectedError:  "Not Found",
	}
	tests.RunUnitTest(t, e, tc)

	// Timeout the last lease
	frozenClock += 3 * time.Hour
	clock.SetAndFreezeClock(t, time.Unix(0, 0).Add(frozenClock))
	tc = tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/reserve",
		Handler:        reserveNode,
		ExpectedStatus: 200,
		ExpectedResult: &models.NodeLease{
			ID:      swag.String("node1"),
			LeaseID: swag.String("2"),
			VpnIP:   ipv4("192.168.100.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)
	actual, err = testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected = map[string]*storage.CINode{
		"node1": {
			Id:            "node1",
			VpnIp:         "192.168.100.1",
			Available:     false,
			LastLeaseTime: timestampProto(t, int64(frozenClock/time.Second)),
		},
	}
	assert.Equal(t, expected, actual)
}

func Test_ReserveSpecificCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	frozenClock := 1000 * time.Hour
	clock.SetAndFreezeClock(t, time.Unix(0, 0).Add(frozenClock))
	defer clock.UnfreezeClock(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci/nodes"

	oHands := handlers.GetObsidianHandlers()
	reserveNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/:node_id/reserve", obsidian.POST).HandlerFunc

	// Empty case
	tc := tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/node1/reserve",
		Handler:        reserveNode,
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		ExpectedStatus: 404,
		ExpectedError:  "Either the node is not known or it has already been reserved.",
	}
	tests.RunUnitTest(t, e, tc)

	// Happy path
	err := testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node1", VpnIP: "192.168.100.1"})
	assert.NoError(t, err)
	tc = tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/:node_id/reserve",
		Handler:        reserveNode,
		ExpectedStatus: 200,
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		ExpectedResult: &models.NodeLease{
			ID:      swag.String("node1"),
			LeaseID: swag.String("manual"),
			VpnIP:   ipv4("192.168.100.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)
	actual, err := testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected := map[string]*storage.CINode{
		"node1": {
			Id:            "node1",
			VpnIp:         "192.168.100.1",
			Available:     false,
			LastLeaseTime: timestampProto(t, int64(frozenClock/time.Second)),
		},
	}
	assert.Equal(t, expected, actual)

	// Pool is empty, manual lease request should override any existing leases
	tc = tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/:node_id/reserve",
		Handler:        reserveNode,
		ParamNames:     []string{"node_id"},
		ParamValues:    []string{"node1"},
		ExpectedStatus: 200,
		ExpectedResult: &models.NodeLease{
			ID:      swag.String("node1"),
			LeaseID: swag.String("manual"),
			VpnIP:   ipv4("192.168.100.1"),
		},
	}
	tests.RunUnitTest(t, e, tc)
}

func Test_ReleaseCINode(t *testing.T) {
	_ = plugin.RegisterPluginForTests(t, &plugin2.FbinternalOrchestratorPlugin{})
	test_init.StartTestService(t)

	frozenClock := 1000 * time.Hour
	clock.SetAndFreezeClock(t, time.Unix(0, 0).Add(frozenClock))
	defer clock.UnfreezeClock(t)

	e := echo.New()
	testURLRoot := "/magma/v1/ci"

	oHands := handlers.GetObsidianHandlers()
	releaseNode := tests.GetHandlerByPathAndMethod(t, oHands, testURLRoot+"/nodes/:node_id/release/:lease_id", obsidian.POST).HandlerFunc

	// Empty case (bad lease/node id)
	tc := tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/node1/release/1",
		ParamNames:     []string{"node_id", "lease_id"},
		ParamValues:    []string{"node1", "1"},
		Handler:        releaseNode,
		ExpectedStatus: 400,
		ExpectedError:  "no node matching the provided ID and lease ID was found",
	}
	tests.RunUnitTest(t, e, tc)

	// Happy path
	err := testcontroller.CreateOrUpdateNode(&storage.MutableCINode{Id: "node1", VpnIP: "192.168.100.1"})
	assert.NoError(t, err)
	actualLease, err := testcontroller.LeaseNode()
	assert.NoError(t, err)
	expectedLease := &storage.NodeLease{Id: "node1", VpnIP: "192.168.100.1", LeaseID: "1"}
	assert.Equal(t, expectedLease, actualLease)
	tc = tests.Test{
		Method:         "POST",
		URL:            testURLRoot + "/nodes/node1/release/1",
		ParamNames:     []string{"node_id", "lease_id"},
		ParamValues:    []string{"node1", "1"},
		Handler:        releaseNode,
		ExpectedStatus: 204,
	}
	tests.RunUnitTest(t, e, tc)

	actual, err := testcontroller.GetNodes(nil)
	assert.NoError(t, err)
	expected := map[string]*storage.CINode{
		"node1": {
			Id:            "node1",
			VpnIp:         "192.168.100.1",
			Available:     true,
			LastLeaseTime: timestampProto(t, int64(frozenClock/time.Second)),
		},
	}
	assert.Equal(t, expected, actual)
}

func expectedDT(t *testing.T, ti time.Duration) strfmt.DateTime {
	zulu, err := time.LoadLocation("")
	assert.NoError(t, err)
	tz := time.Unix(0, 0).Add(ti).In(zulu)
	return strfmt.DateTime(tz)
}

func ipv4(s string) *strfmt.IPv4 {
	ip := strfmt.IPv4(s)
	return &ip
}

func timestampProto(t *testing.T, ti int64) *timestamp.Timestamp {
	ret, err := ptypes.TimestampProto(time.Unix(ti, 0))
	assert.NoError(t, err)
	return ret
}