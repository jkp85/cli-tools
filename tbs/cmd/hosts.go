package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/hosts"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	cmd := hostsCmd()
	cmd.AddCommand(
		hostListCmd(),
		hostCreateCmd(),
		hostUpdateCmd(),
		hostDeleteCmd(),
	)
	RootCmd.AddCommand(cmd)
}

func hostsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host",
		Short: "Handle your hosts",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("host_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func hostListCmd() *cobra.Command {
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "Host list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := hosts.NewHostsListParams()
			lf.Apply(params)
			resp, err := cli.Hosts.HostsList(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("host_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	return cmd
}

func hostCreateCmd() *cobra.Command {
	body := &models.DockerHostData{
		Name: new(string),
		IP:   new(string),
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if *body.Name == "" {
				return errors.New("You need to provide name for your host")
			}
			cli := api.Client()
			params := hosts.NewHostsCreateParams()
			params.SetNamespace(cli.Namespace)
			params.SetData(body)
			resp, err := cli.Hosts.HostsCreate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("host_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Name, "name", "", "Host name")
	cmd.Flags().StringVar(body.IP, "ip", "", "Host ip")
	cmd.Flags().Int64Var(&body.Port, "port", 0, "Host port")
	return cmd
}

func hostUpdateCmd() *cobra.Command {
	var hostID string
	body := &models.DockerHostData{
		Name: new(string),
		IP:   new(string),
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update host",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if *body.Name == "" && hostID == "" {
				return errors.New("You must provide either host name or id")
			}
			cli := api.Client()
			if hostID == "" {
				hostID, err = cli.GetHostIDByName(*body.Name)
			}
			params := hosts.NewHostsUpdateParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(hostID)
			params.SetData(body)
			resp, err := cli.Hosts.HostsUpdate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("host_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&hostID, "uuid", "", "Host id")
	cmd.Flags().StringVar(body.Name, "name", "", "Host name")
	cmd.Flags().StringVar(body.IP, "ip", "", "Host ip")
	cmd.Flags().Int64Var(&body.Port, "port", 0, "Host port")
	return cmd
}

func hostDeleteCmd() *cobra.Command {
	var hostID, hostName string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hostName == "" && hostID == "" {
				return errors.New("You must provide host name or id")
			}
			confirm, err := readStdin(
				fmt.Sprintf("Are you sure you want to delete host '%s'? (Y/n)", hostName))
			if err != nil {
				return err
			}
			confirm = strings.ToLower(confirm)
			if confirm == "n" || confirm == "no" {
				jww.FEEDBACK.Println("Aborted")
				return nil
			}
			cli := api.Client()
			if hostID == "" {
				hostID, err = cli.GetHostIDByName(hostName)
				if err != nil {
					return err
				}
			}
			params := hosts.NewHostsDeleteParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(hostID)
			_, err = cli.Hosts.HostsDelete(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Host deleted")
			return nil
		},
	}
	return cmd
}
