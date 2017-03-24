package api

import (
	"net/url"

	apiclient "github.com/3Blades/go-sdk/client"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func Client() *apiclient.Threeblades {
	apiRoot := viper.GetString("root")
	root, err := url.Parse(apiRoot)
	if err != nil {
		jww.ERROR.Fatal("You need to provide valid url as api root.")
		return nil
	}
	token := viper.GetString("token")
	transport := httptransport.New(root.Host, "", []string{root.Scheme})
	if token != "" {
		transport.DefaultAuthentication = httptransport.APIKeyAuth("AUTHORIZATION", "header", "Token "+token)
	}
	return apiclient.New(transport, strfmt.Default)
}
