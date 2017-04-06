package cmd

import (
	"errors"
	"fmt"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	sCmd := serverCmd()
	sCmd.AddCommand(serverLsCmd())
	sCmd.AddCommand(serverCreateCmd())
	sCmd.AddCommand(serverUpdateCmd())
	sCmd.AddCommand(serverDescribeCmd())
	sCmd.AddCommand(serverStartCmd())
	sCmd.AddCommand(serverStopCmd())
	sCmd.AddCommand(serverTerminateCmd())
	RootCmd.AddCommand(sCmd)
}

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "User server management",
	}
	cmd.Flags().String("format", "json", "Output format")
	viper.BindPFlag("server_format", cmd.Flags().Lookup("format"))
	return cmd
}

func serverLsCmd() *cobra.Command {
	ls := utils.ListFlags{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersListParams()
			ls.Apply(params)
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			params.SetProjectPk(projectID)
			resp, err := cli.Projects.ProjectsServersList(params)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	ls.Set(cmd)
	return cmd
}

type ServerConfig struct {
	Script   string `json:"script,omitempty"`
	Function string `json:"function,omitempty"`
	Command  string `json:"command,omitempty"`
	Type     string `json:"type,omitempty"`
}

func serverCreateCmd() *cobra.Command {
	body := projects.ProjectsServersCreateBody{
		Name:                 new(string),
		EnvironmentResources: new(string),
		Connected:            []string{},
	}
	bodyConf := &ServerConfig{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			body.Config = bodyConf
			params.SetData(body)
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			params.SetProjectPk(projectID)
			resp, err := cli.Projects.ProjectsServersCreate(params)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&body.ImageName, "image", "", "Server image")
	cmd.Flags().StringVar(body.EnvironmentResources, "resources", "", "Server resources")
	cmd.Flags().StringVar(&body.StartupScript, "startup-script", "", "Server startup script")
	cmd.Flags().StringVar(&bodyConf.Function, "function", "", "Function to run")
	cmd.Flags().StringVar(&bodyConf.Script, "script", "", "Script to run")
	cmd.Flags().StringVar(&bodyConf.Command, "command", "", "Command to run")
	cmd.Flags().StringVar(&bodyConf.Type, "type", "", "Server type [restful,cron,jupyter]")
	return cmd
}

func getServerByName(name, projectID string) (*models.Server, error) {
	cli := api.Client()
	params := projects.NewProjectsServersListParams()
	ns := viper.GetString("namespace")
	params.SetNamespace(ns)
	params.SetProjectPk(projectID)
	params.SetName(&name)
	resp, err := cli.Projects.ProjectsServersList(params)
	if err != nil {
		return &models.Server{}, err
	}
	if len(resp.Payload) < 1 {
		return &models.Server{}, fmt.Errorf("There is no server with name: %s", name)
	}
	return resp.Payload[0], nil
}

func getServerByID(serverID, projectID string) (*models.Server, error) {
	cli := api.Client()
	params := projects.NewProjectsServersReadParams()
	ns := viper.GetString("namespace")
	params.SetNamespace(ns)
	params.SetID(serverID)
	params.SetProjectPk(projectID)
	resp, err := cli.Projects.ProjectsServersRead(params)
	if err != nil {
		return &models.Server{}, err
	}
	return resp.Payload, nil
}

func serverDescribeCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Server details",
		RunE: func(cmd *cobra.Command, args []string) error {
			var server *models.Server
			var err error
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			if name == "" && serverID == "" {
				return errors.New("You must specify either name or id")
			}
			if serverID != "" {
				server, err = getServerByID(serverID, projectID)
			} else {
				server, err = getServerByName(name, projectID)
			}
			if err != nil {
				return err
			}
			return api.Render("server_format", server)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}

func serverUpdateCmd() *cobra.Command {
	var serverID string
	body := projects.ProjectsServersPartialUpdateBody{
		Connected: []string{},
	}
	bodyConf := &ServerConfig{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update server",
		RunE: func(cmd *cobra.Command, args []string) error {
			body.Config = bodyConf
			cli := api.Client()
			params := projects.NewProjectsServersPartialUpdateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			if serverID == "" {
				server, err := getServerByName(body.Name, projectID)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			params.SetID(serverID)
			params.SetData(body)
			params.SetProjectPk(projectID)
			resp, err := cli.Projects.ProjectsServersPartialUpdate(params)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	cmd.Flags().StringVar(&body.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&body.ImageName, "image", "", "Server image")
	cmd.Flags().StringVar(&body.EnvironmentResources, "resources", "", "Server resources")
	cmd.Flags().StringVar(&body.StartupScript, "startup-script", "", "Server startup script")
	cmd.Flags().StringVar(&bodyConf.Function, "function", "", "Function to run")
	cmd.Flags().StringVar(&bodyConf.Script, "script", "", "Script to run")
	cmd.Flags().StringVar(&bodyConf.Command, "command", "", "Command to run")
	cmd.Flags().StringVar(&bodyConf.Type, "type", "", "Server type [restful,cron,jupyter]")
	return cmd
}

func serverStartCmd() *cobra.Command {
	var serverID, name string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			if serverID == "" {
				server, err := getServerByName(name, projectID)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			cli := api.Client()
			params := projects.NewProjectsServersStartCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetProjectPk(projectID)
			params.SetServerPk(serverID)
			_, err = cli.Projects.ProjectsServersStartCreate(params)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Server started")
			return nil
		},
	}
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	return cmd
}

func serverStopCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			if serverID == "" {
				server, err := getServerByName(name, projectID)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			cli := api.Client()
			params := projects.NewProjectsServersStopCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetProjectPk(projectID)
			params.SetServerPk(serverID)
			_, err = cli.Projects.ProjectsServersStopCreate(params)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Server stopped")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}

func serverTerminateCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			projectID, err := getProjectIDByName(viper.GetString("project"))
			if err != nil {
				return err
			}
			if serverID == "" {
				server, err := getServerByName(name, projectID)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			cli := api.Client()
			params := projects.NewProjectsServersTerminateCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetProjectPk(projectID)
			params.SetServerPk(serverID)
			_, err = cli.Projects.ProjectsServersTerminateCreate(params)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Server terminated")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}
