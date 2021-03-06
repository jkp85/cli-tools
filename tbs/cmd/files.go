package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
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
	fCmd.AddCommand(fileDeleteCmd())
	fCmd.AddCommand(fileUploadCmd())
	RootCmd.AddCommand(fCmd)
}

func fileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File management",
	}
	// cmd.PersistentFlags().String("format", "json", "Output format")
	// viper.BindPFlag("file_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func fileListCommand() *cobra.Command {
	ls := utils.ListFlags{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsProjectFilesListParams()
			ls.Apply(params)
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProject(projectID)
			resp, err := cli.Projects.ProjectsProjectFilesList(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("file_format", resp.Payload)
		},
	}
	ls.Set(cmd)
	return cmd
}

func getFileByName(name, projectID string) (*models.ProjectFile, error) {
	cli := api.Client()
	params := projects.NewProjectsProjectFilesListParams()
	params.SetNamespace(cli.Namespace)
	params.SetProject(projectID)
	resp, err := cli.Projects.ProjectsProjectFilesList(params, cli.AuthInfo)
	if err != nil {
		return &models.ProjectFile{}, err
	}
	if len(resp.Payload) < 1 {
		return &models.ProjectFile{}, fmt.Errorf("There is no file with name/path: %s", name)
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
			params := projects.NewProjectsProjectFilesDeleteParams()
			params.SetNamespace(cli.Namespace)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProject(projectID)
			for _, arg := range args {
				if !utils.IsUUID(arg) {
					file, err := getFileByName(arg, projectID)
					if err != nil {
						jww.FEEDBACK.Printf("There is no file with name %s", arg)
					}
					arg = file.ID
				}
				params.SetID(arg)
				_, err = cli.Projects.ProjectsProjectFilesDelete(params, cli.AuthInfo)
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

func newFileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if paramName != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			jww.ERROR.Printf("There was an error resolving path: %s\n", path)
			return nil, err
		}

		localFile, err := os.Open(abs)
		if err != nil {
			jww.ERROR.Printf("There was an error opening file: %s\n", path)
			return nil, err
		}
		defer localFile.Close()

		part, err := writer.CreateFormFile(paramName, filepath.Base(abs))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, localFile)

		for key, val := range params {
			_ = writer.WriteField(key, val)

		}
		err = writer.Close()
		if err != nil {
			return nil, err
		}
	} else {
		jsonValue, _ := json.Marshal(params)
		body = bytes.NewBuffer(jsonValue)
	}

	req, err := http.NewRequest("POST", uri, body)
	if paramName != "" {
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	token := viper.GetString("token")
	req.Header.Set("AUTHORIZATION", "JWT "+token)

	return req, err
}

func getFileUploadResponse(request *http.Request) (*bytes.Buffer, error) {
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()
	return body, nil
}

func fileUploadCmd() *cobra.Command {
	uploadBody := &models.ProjectFile{}
	cmd := &cobra.Command{
		Use:   "upload [files]",
		Short: "Upload files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			projectID, err := cli.GetProjectID()

			if err != nil {
				return err
			}

			rootUrl := viper.GetString("root")
			endPoint := fmt.Sprintf("/%v/projects/%v/project_files/", cli.Namespace, projectID)
			apiUrl := rootUrl + endPoint

			extraParams := map[string]string{
				"project":     projectID,
				"name":        uploadBody.Name,
				"base64_data": uploadBody.Content,
			}
			if len(args) > 0 {
				for _, path := range args {
					request, err := newFileUploadRequest(apiUrl, extraParams, "file", path)

					if err != nil {
						jww.ERROR.Printf("There was an error uploading file: %s\n", path)
						continue
					}

					body, err := getFileUploadResponse(request)
					if err != nil {
						jww.ERROR.Printf("There was an error uploading file: %s\n", path)
					}
					fmt.Println(body)
				}
			} else {
				request, err := newFileUploadRequest(apiUrl, extraParams, "", "")
				if err != nil {
					jww.ERROR.Printf("There was an error uploading file: %s\n", uploadBody.Name)
				}

				body, err := getFileUploadResponse(request)

				if err != nil {
					jww.ERROR.Printf("There was an error uploading file: %s\n", uploadBody.Name)
				}

				fmt.Println(body)
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&uploadBody.Name, "name", "", "The file's name, only to be used in conjunction with base64 data")
	flags.StringVar(&uploadBody.Content, "content", "", "Content as base64 encoded stgring.")
	return cmd
}
