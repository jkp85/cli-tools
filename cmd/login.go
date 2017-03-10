// Copyright © 2017 3Blades LLC
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/jkp85/cli-tools/api"
	"github.com/jkp85/go-sdk/client/auth"
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
			token, err := getToken("localhost:5000", username, password)
			if err != nil {
				return err
			}
			viper.Set("token", token)
			return saveToken(token)
		},
	}
	flags := loginCmd.Flags()
	flags.StringVarP(&username, "username", "u", "", "Username")
	flags.StringVarP(&password, "password", "p", "", "Password")
	return loginCmd
}

func init() {
	RootCmd.AddCommand(newLoginCmd())
	token, err := ioutil.ReadFile(tokenFilePath())
	if err == nil {
		viper.Set("token", token)
	}
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

func getToken(server, username, password string) (string, error) {
	cli := api.APIClient()
	params := auth.NewAuthSimpleTokenAuthCreateParams()
	params.SetData(auth.AuthSimpleTokenAuthCreateBody{
		Username: &username,
		Password: &password,
	})
	resp, err := cli.Auth.AuthSimpleTokenAuthCreate(params)
	if err != nil {
		return "", err
	}
	return resp.Payload.Token, nil
}

func tokenFilePath() string {
	configFilePath := viper.ConfigFileUsed()
	return filepath.Join(filepath.Dir(configFilePath), ".threeblades.token")
}

func saveToken(token string) error {
	return ioutil.WriteFile(tokenFilePath(), []byte(token), 0600)
}
