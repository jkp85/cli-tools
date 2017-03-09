package api

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/jkp85/go-sdk/client"
	"github.com/spf13/viper"
)

func APIClient() *apiclient.Threeblades {
	apiRoot := viper.GetString("root")
	token := viper.GetString("token")
	transport := httptransport.New(apiRoot, "", []string{"http"})
	if token != "" {
		transport.DefaultAuthentication = httptransport.APIKeyAuth("AUTHORIZATION", "header", "Token "+token)
	}
	return apiclient.New(transport, strfmt.Default)
}
