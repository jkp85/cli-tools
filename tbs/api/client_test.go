package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/3Blades/cli-tools/tbs/utils"
	apiclient "github.com/3Blades/go-sdk/client"
	"github.com/3Blades/go-sdk/models"
	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
)

func runServer(data interface{}) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestGetProjectIDByName(t *testing.T) {
	name := "Test"
	project := &models.Project{
		Name: &name,
		ID:   uuid.NewV4().String(),
	}
	server := runServer([]*models.Project{project})
	defer server.Close()
	cli := &APIClient{
		apiclient.New(transport(server.URL, ""), strfmt.Default),
		"", "", "", "", "",
	}
	id, err := cli.GetProjectIDByName(name)
	if err != nil {
		t.Error(err)
	}
	if project.ID != id {
		t.Error("Wrong project id.")
	}
}

func TestListServers(t *testing.T) {
	projectName := "Test"
	projectID := uuid.NewV4().String()
	servers := []*models.Server{
		{
			ID:        uuid.NewV4().String(),
			Name:      new(string),
			Connected: []string{},
		},
		{
			ID:        uuid.NewV4().String(),
			Name:      new(string),
			Connected: []string{},
		},
	}
	server := runServer(servers)
	defer server.Close()
	cli := &APIClient{
		apiclient.New(transport(server.URL, ""), strfmt.Default),
		"test", projectName, projectID, "", "",
	}
	results, err := cli.ListServers(&utils.ListFlags{})
	if err != nil {
		t.Error(err)
	}
	for i, result := range results {
		if servers[i].ID != result.ID {
			t.Error("Wrong IDs")
		}
	}
}

func TestGetServerByName(t *testing.T) {
	serverName := "Test"
	apiServer := &models.Server{
		Name:      &serverName,
		ID:        uuid.NewV4().String(),
		Connected: []string{},
	}
	server := runServer([]*models.Server{apiServer})
	defer server.Close()
	cli := &APIClient{
		apiclient.New(transport(server.URL, ""), strfmt.Default),
		"test", "Test", uuid.NewV4().String(), "", "",
	}
	id, err := cli.GetServerByName(serverName)
	if err != nil {
		t.Fatal(err)
	}
	if apiServer.ID != id.ID {
		t.Error("Servers ids don't match")
	}
}

func TestGetServerByID(t *testing.T) {
	serverName := "Test"
	serverID := uuid.NewV4().String()
	apiServer := &models.Server{
		Name:      &serverName,
		ID:        serverID,
		Connected: []string{},
	}
	server := runServer(apiServer)
	defer server.Close()
	cli := &APIClient{
		apiclient.New(transport(server.URL, ""), strfmt.Default),
		"test", "Test", uuid.NewV4().String(), "", "",
	}
	result, err := cli.GetServerByID(serverID)
	if err != nil {
		t.Error(err)
	}
	if *result.Name != serverName {
		t.Error("Server names don't match")
	}
}
