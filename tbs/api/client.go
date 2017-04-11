package api

import (
	"fmt"
	"net/url"

	"github.com/3Blades/cli-tools/tbs/utils"
	apiclient "github.com/3Blades/go-sdk/client"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
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
}

func (c *APIClient) GetProjectIDByName(name string) (string, error) {
	if c.projectID != "" {
		return c.projectID, nil
	}
	params := projects.NewProjectsListParams()
	params.SetNamespace(c.Namespace)
	params.SetName(&name)
	resp, err := c.Projects.ProjectsList(params)
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
	params.SetProjectPk(projectID)
	resp, err := c.Projects.ProjectsServersList(params)
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
	params.SetProjectPk(projectID)
	params.SetName(&name)
	resp, err := c.Projects.ProjectsServersList(params)
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
	params.SetProjectPk(projectID)
	resp, err := c.Projects.ProjectsServersRead(params)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func Client() *APIClient {
	cli := apiclient.New(transport(viper.GetString("root"), viper.GetString("token")), strfmt.Default)
	return &APIClient{
		cli,
		viper.GetString("namespace"),
		viper.GetString("project"),
		viper.GetString("projectID"),
		viper.GetString("server"),
		viper.GetString("serverID"),
	}
}

func transport(apiRoot, token string) *httptransport.Runtime {
	root, err := url.Parse(apiRoot)
	if err != nil {
		jww.ERROR.Fatal("You need to provide valid url as api root.")
		return nil
	}
	tr := httptransport.New(root.Host, "", []string{root.Scheme})
	if token != "" {
		tr.DefaultAuthentication = httptransport.APIKeyAuth("AUTHORIZATION", "header", "Token "+token)
	}
	return tr
}
