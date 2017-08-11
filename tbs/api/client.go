package api

import (
	"fmt"
	"net/url"

	"github.com/3Blades/cli-tools/tbs/utils"
	apiclient "github.com/3Blades/go-sdk/client"
	"github.com/3Blades/go-sdk/client/hosts"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

type APIClient struct {
	*apiclient.Threeblades
	Namespace string
	project   string
	projectID string
	server    string
	serverID  string
	AuthInfo  runtime.ClientAuthInfoWriterFunc
}

func (c *APIClient) GetProjectIDByName(name string) (string, error) {
	if c.projectID != "" {
		return c.projectID, nil
	}
	params := projects.NewProjectsListParams()
	params.SetNamespace(c.Namespace)
	params.SetName(&name)
	resp, err := c.Projects.ProjectsList(params, c.AuthInfo)
	if err != nil {
		return "", err
	}
	if len(resp.Payload) == 0 {
		return "", fmt.Errorf("There is no project with name: '%s'", name)
	}
	return resp.Payload[0].ID, nil
}

func (c *APIClient) GetProjectID() (string, error) {
	if c.projectID != "" {
		return c.projectID, nil
	}
	if c.project == "" {
		return "", fmt.Errorf("Project name is blank. Please set project name with env command.")
	}
	var err error
	c.projectID, err = c.GetProjectIDByName(c.project)
	return c.projectID, err
}

func (c *APIClient) ListServers(ls *utils.ListFlags) ([]*models.Server, error) {
	params := projects.NewProjectsServersListParams()
	ls.Apply(params)
	params.SetNamespace(c.Namespace)
	projectID, err := c.GetProjectID()
	if err != nil {
		return []*models.Server{}, err
	}
	params.SetProjectID(projectID)
	resp, err := c.Projects.ProjectsServersList(params, c.AuthInfo)
	if err != nil {
		return []*models.Server{}, err
	}
	return resp.Payload, nil
}

func (c *APIClient) GetServerByName(name string) (*models.Server, error) {
	params := projects.NewProjectsServersListParams()
	params.SetNamespace(c.Namespace)
	projectID, err := c.GetProjectID()
	if err != nil {
		return nil, err
	}
	params.SetProjectID(projectID)
	params.SetName(&name)
	resp, err := c.Projects.ProjectsServersList(params, c.AuthInfo)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) < 1 {
		return nil, fmt.Errorf("There is no server with name: %s", name)
	}
	return resp.Payload[0], nil
}

func (c *APIClient) GetServerByID(serverID string) (*models.Server, error) {
	params := projects.NewProjectsServersReadParams()
	params.SetNamespace(c.Namespace)
	params.SetID(serverID)
	projectID, err := c.GetProjectID()
	if err != nil {
		return nil, err
	}
	params.SetProjectID(projectID)
	resp, err := c.Projects.ProjectsServersRead(params, c.AuthInfo)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (c *APIClient) GetHostIDByName(hostName string) (string, error) {
	params := hosts.NewHostsListParams()
	params.SetNamespace(c.Namespace)
	params.SetName(&hostName)
	resp, err := c.Hosts.HostsList(params, c.AuthInfo)
	if err != nil {
		return "", err
	}
	if len(resp.Payload) < 1 {
		return "", fmt.Errorf("There is no host with name: %s", hostName)
	}
	return resp.Payload[0].ID, nil
}

func (c *APIClient) GetServerTriggerByName(projectID, serverID, name string) (*models.ServerAction, error) {
	params := projects.NewServiceTriggerListParams()
	params.SetNamespace(c.Namespace)
	params.SetProjectID(projectID)
	params.SetServerID(serverID)
	params.SetName(&name)
	resp, err := c.Projects.ServiceTriggerList(params, c.AuthInfo)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) < 1 {
		return nil, fmt.Errorf("There is no trigger with name: %s", name)
	}
	return resp.Payload[0], nil
}

func (c *APIClient) GetServerTriggerByID(projectID, serverID, ID string) (*models.ServerAction, error) {
	params := projects.NewServiceTriggerReadParams()
	params.SetNamespace(c.Namespace)
	params.SetProjectID(projectID)
	params.SetServerID(serverID)
	params.SetID(ID)
	resp, err := c.Projects.ServiceTriggerRead(params, c.AuthInfo)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func Client() *APIClient {
	cli := apiclient.New(transport(viper.GetString("root")), strfmt.Default)
	return &APIClient{
		cli,
		viper.GetString("namespace"),
		viper.GetString("project"),
		viper.GetString("projectID"),
		viper.GetString("server"),
		viper.GetString("serverID"),
		AuthInfo,
	}
}

func AuthInfo(req runtime.ClientRequest, reg strfmt.Registry) error {
	return req.SetHeaderParam("AUTHORIZATION", fmt.Sprintf("Bearer %s", viper.GetString("token")))
}

func transport(apiRoot string) *httptransport.Runtime {
	root, err := url.Parse(apiRoot)
	if err != nil {
		jww.FATAL.Fatal(err)
	}
	return httptransport.New(root.Host, "", []string{root.Scheme})
}
