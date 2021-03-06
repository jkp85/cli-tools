package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/go-sdk/client/auth"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func newLoginCmd() *cobra.Command {
	var username string
	var password string
	loginCmd := &cobra.Command{
		Use:   "login [server]",
		Short: "Login to 3Blades",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetString("token") != "" {
				return nil
			}
			var err error
			if username == "" {
				username, err = readStdin("Username: ")
				if err != nil {
					return err
				}
			}
			if password == "" {
				password, err = readPassword()
				if err != nil {
					return err
				}
			}
			token, err := getToken(username, password)
			if err != nil {
				return err
			}
			viper.Set("token", token)
			err = saveToken(token)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("\nLogin successful")
			return nil
		},
	}
	flags := loginCmd.Flags()
	flags.StringVarP(&username, "username", "u", "", "Username")
	flags.StringVarP(&password, "password", "p", "", "Password")
	return loginCmd
}

func init() {
	RootCmd.AddCommand(newLoginCmd())
}

func readStdin(promptMsg string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	jww.FEEDBACK.Printf(promptMsg)
	out, err := reader.ReadString('\n')
	return strings.TrimSpace(out), err
}

func readPassword() (string, error) {
	jww.FEEDBACK.Printf("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	return strings.TrimSpace(string(bytePassword)), err
}

func getToken(username, password string) (string, error) {
	cli := api.Client()
	params := auth.NewAuthJwtTokenAuthParams()
	params.SetJwtData(&models.JWTData{
		Username: &username,
		Password: &password,
	})
	resp, err := cli.Auth.AuthJwtTokenAuth(params)
	if err != nil {
		return "", err
	}
	return resp.Payload.Token, nil
}

func tokenFilePath() string {
	return filepath.Join(filepath.Dir(viper.ConfigFileUsed()), ".threeblades.token")
}

func saveToken(token string) error {
	return ioutil.WriteFile(tokenFilePath(), []byte(token), 0600)
}
