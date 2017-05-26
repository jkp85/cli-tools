package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	fCmd := fileCmd()
	fCmd.AddCommand(fileListCommand())
	fCmd.AddCommand(fileCreateCmd())
	fCmd.AddCommand(fileDeleteCmd())
	fCmd.AddCommand(fileUploadCmd())
	RootCmd.AddCommand(fCmd)
}

func fileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File management",
	}
	cmd.PersistentFlags().String("format", "json", "Output format")
	viper.BindPFlag("file_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func fileListCommand() *cobra.Command {
	ls := utils.ListFlags{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsFilesListParams()
			ls.Apply(params)
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectPk(projectID)
			resp, err := cli.Projects.ProjectsFilesList(params)
			if err != nil {
				return err
			}
			return api.Render("file_format", resp.Payload)
		},
	}
	ls.Set(cmd)
	return cmd
}

func fileCreateCmd() *cobra.Command {
	var content string
	body := projects.ProjectsFilesCreateBody{
		Path: new(string),
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			user, err := getUserByName(cli.Namespace)
			if err != nil {
				return err
			}
			body.Project = &projectID
			body.Author = &user.ID
			encoded := base64.StdEncoding.EncodeToString([]byte(content))
			body.Content = &encoded
			encoding := "utf-8"
			body.Encoding = &encoding
			params := projects.NewProjectsFilesCreateParams()
			params.SetNamespace(cli.Namespace)
			params.SetProjectPk(projectID)
			params.SetData(body)
			resp, err := cli.Projects.ProjectsFilesCreate(params)
			if err != nil {
				return err
			}
			return api.Render("file_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Path, "name", "", "File name/path")
	cmd.Flags().StringVar(&content, "content", "", "File contents")
	return cmd
}

func getFileByName(name, projectID string) (*models.File, error) {
	cli := api.Client()
	params := projects.NewProjectsFilesListParams()
	params.SetNamespace(cli.Namespace)
	params.SetProjectPk(projectID)
	params.SetPath(&name)
	resp, err := cli.Projects.ProjectsFilesList(params)
	if err != nil {
		return &models.File{}, err
	}
	if len(resp.Payload) < 1 {
		return &models.File{}, fmt.Errorf("There is no file with name/path: %s", name)
	}
	return resp.Payload[0], nil
}

func fileDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [names or ids...]",
		Short: "Delete file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			if len(args) == 0 {
				return errors.New("You must provide at least one name or id")
			}
			params := projects.NewProjectsFilesDeleteParams()
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectPk(projectID)
			for _, arg := range args {
				if !utils.IsUUID(arg) {
					file, err := getFileByName(arg, projectID)
					if err != nil {
						jww.FEEDBACK.Printf("There is no file with name %s", arg)
					}
					arg = file.ID
				}
				params.SetID(arg)
				_, err = cli.Projects.ProjectsFilesDelete(params)
				if err != nil {
					jww.FEEDBACK.Println(err)
				}
				jww.FEEDBACK.Printf("File %s deleted\n", arg)
			}
			return nil
		},
	}
	return cmd
}

func fileUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload [files]",
		Short: "Upload files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			user, err := getUserByName(cli.Namespace)
			if err != nil {
				return err
			}
			encoding := "utf-8"
			for _, path := range args {
				abs, err := filepath.Abs(path)
				if err != nil {
					jww.ERROR.Printf("There was an error resolving path: %s\n", path)
					continue
				}
				body := projects.ProjectsFilesCreateBody{
					Path:     &path,
					Author:   &user.ID,
					Encoding: &encoding,
					Project:  &projectID,
				}
				contentB, err := ioutil.ReadFile(abs)
				if err != nil {
					jww.ERROR.Printf("There was an error opening file: %s\n", path)
					continue
				}
				encoded := base64.StdEncoding.EncodeToString(contentB)
				body.Content = &encoded
				params := projects.NewProjectsFilesCreateParams()
				params.SetNamespace(cli.Namespace)
				params.SetProjectPk(projectID)
				params.SetData(body)
				_, err = cli.Projects.ProjectsFilesCreate(params)
				if err != nil {
					jww.ERROR.Printf("There was an error uploading file: %s\n", path)
					continue
				}
				jww.FEEDBACK.Printf("File uploaded: %s\n", path)
			}
			return nil
		},
	}
	return cmd
}
