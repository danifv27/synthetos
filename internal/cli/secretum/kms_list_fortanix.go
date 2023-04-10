package secretum

import (
	"fry.org/cmo/cli/internal/cli/common"
	"github.com/workanator/go-floc/v3"
)

type KmsListFortanixCmd struct {
	Flags  KmsListFortanixFlags     `embed:""`
	Groups KmsListFortanixGroupsCmd `cmd:"" help:"List Fortanix groups."`
}

type KmsListFortanixFlags struct {
	ApiEndpointURL string `help:"The URL for the Fortanix API endpoint. Make sure to include the trailing slash." prefix:"kms.list.fortanix." env:"SC_KMS_LIST_FORTANIX_API_ENDPOINT_URL" default:"https://api.fortanix.com/"`
	AccessKey      string `help:"Your Fortanix API access key. You can obtain this key by logging into your Fortanix account and navigating to the 'API Keys' page in the 'Settings' section." prefix:"kms.list.fortanix." env:"SC_KMS_LIST_FORTANIX_ACCESS_KEY" required:""`
	SecretKey      string `help:"Your Fortanix API secret key. This key is displayed only once when you create the key." prefix:"kms.list.fortanix." env:"SC_KMS_LIST_FORTANIX_SECRET_KEY" required:""`
}

func initializeKmsListFortanixCmd(ctx floc.Context, ctrl floc.Control) error {

	return nil
}

func (cmd *KmsListFortanixCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append(c.InitSeq, initializeKmsListFortanixCmd)

	return nil
}
