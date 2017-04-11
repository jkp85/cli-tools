package cmd

import (
	"errors"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	cmd := serverCmd()
	cmd.AddCommand(
		serverLsCmd(),
		serverCreateCmd(),
		serverUpdateCmd(),
		serverDescribeCmd(),
		serverStartCmd(),
		serverStopCmd(),
		serverTerminateCmd(),
	)
	RootCmd.AddCommand(cmd)
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
	ls := &utils.ListFlags{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			resp, err := cli.ListServers(ls)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp)
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
			params.SetNamespace(cli.Namespace)
			body.Config = bodyConf
			params.SetData(body)
			projectID, err := cli.GetProjectID()
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

func serverDescribeCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Server details",
		RunE: func(cmd *cobra.Command, args []string) error {
			var server *models.Server
			var err error
			cli := api.Client()
			if name == "" && serverID == "" {
				return errors.New("You must specify either name or id")
			}
			if serverID != "" {
				server, err = cli.GetServerByID(serverID)
			} else {
				server, err = cli.GetServerByName(name)
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
			params.SetNamespace(cli.Namespace)
			if serverID == "" {
				server, err := cli.GetServerByName(body.Name)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			params.SetID(serverID)
			params.SetData(body)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
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
			cli := api.Client()
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			if serverID == "" {
				server, err := cli.GetServerByName(name)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			params := projects.NewProjectsServersStartCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
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
			cli := api.Client()
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			if serverID == "" {
				server, err := cli.GetServerByName(name)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			params := projects.NewProjectsServersStopCreateParams()
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
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
			cli := api.Client()
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			if serverID == "" {
				server, err := cli.GetServerByName(name)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			params := projects.NewProjectsServersTerminateCreateParams()
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
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
