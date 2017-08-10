package cmd

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/gorilla/websocket"
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
		serverLogsCmd(),
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

func serverCreateCmd() *cobra.Command {
	body := &models.ServerData{
		Name:      new(string),
		Connected: []string{},
	}
	bodyConf := &models.ServerConfig{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersCreateParams()
			params.SetNamespace(cli.Namespace)
			body.Config = bodyConf
			params.SetServerData(body)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectID(projectID)
			resp, err := cli.Projects.ProjectsServersCreate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&body.ImageName, "image", "", "Server image")
	cmd.Flags().StringVar(&body.ServerSize, "resources", "", "Server resources")
	cmd.Flags().StringVar(&body.StartupScript, "startup-script", "", "Server startup script")
	cmd.Flags().StringVar(&bodyConf.Function, "function", "", "Function to run")
	cmd.Flags().StringVar(&bodyConf.Script, "script", "", "Script to run")
	cmd.Flags().StringVar(&bodyConf.Command, "command", "", "Command to run")
	cmd.Flags().StringVar(&bodyConf.Type, "type", "", "Server type [restful,cron,jupyter]")
	cmd.Flags().StringVar(&body.Host, "host", "", "Host id to run server on.")
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
	body := &models.ServerData{
		Connected: []string{},
		Name:      new(string),
	}
	bodyConf := &models.ServerConfig{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update server",
		RunE: func(cmd *cobra.Command, args []string) error {
			body.Config = bodyConf
			cli := api.Client()
			params := projects.NewProjectsServersUpdateParams()
			params.SetNamespace(cli.Namespace)
			if serverID == "" {
				server, err := cli.GetServerByName(*body.Name)
				if err != nil {
					return err
				}
				serverID = server.ID
			}
			params.SetID(serverID)
			params.SetServerData(body)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectID(projectID)
			resp, err := cli.Projects.ProjectsServersUpdate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	cmd.Flags().StringVar(body.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&body.ImageName, "image", "", "Server image")
	cmd.Flags().StringVar(&body.StartupScript, "resources", "", "Server resources")
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
			params := projects.NewProjectsServersStartParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectID(projectID)
			params.SetID(serverID)
			_, err = cli.Projects.ProjectsServersStart(params, cli.AuthInfo)
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
			params := projects.NewProjectsServersStopParams()
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectID(projectID)
			params.SetID(serverID)
			_, err = cli.Projects.ProjectsServersStop(params, cli.AuthInfo)
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
			params := projects.NewProjectsServersTerminateParams()
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectID(projectID)
			params.SetID(serverID)
			_, err = cli.Projects.ProjectsServersTerminate(params, cli.AuthInfo)
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

func serverLogsCmd() *cobra.Command {
	var name, serverID, logsURL string
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Server logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			cli := api.Client()
			var server *models.Server
			var err error
			if serverID == "" {
				server, err = cli.GetServerByName(name)
			} else {
				server, err = cli.GetServerByID(serverID)
			}
			if err != nil {
				return err
			}
			serverID = server.ID
			logsURL = server.LogsURL
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)
			ws, err := url.Parse(logsURL)
			if err != nil {
				return err
			}
			header := make(http.Header)
			header.Add("Origin", viper.GetString("root"))
			c, _, err := websocket.DefaultDialer.Dial(ws.String(), header)
			if err != nil {
				return err
			}
			defer c.Close()
			done := make(chan struct{})
			go func() {
				defer c.Close()
				defer close(done)

				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						return
					}
					jww.FEEDBACK.Println(string(message))
				}
			}()
			for {
				select {
				case <-interrupt:
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					c.Close()
					return nil
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}
