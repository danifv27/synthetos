package secretum

import (
	"fmt"
	"net/url"

	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

type KmsListFortanixCmd struct {
	Flags  KmsListFortanixFlags     `embed:""`
	Groups KmsListFortanixGroupsCmd `cmd:"" help:"List Fortanix groups."`
}

type KmsListFortanixFlags struct {
	ApiEndpointURL string `help:"The URL for the Fortanix API endpoint. Make sure to include the trailing slash." prefix:"kms.list.fortanix." env:"SC_KMS_LIST_FORTANIX_API_ENDPOINT_URL" default:"https://api.fortanix.com"`
	ApiKey         string `help:"Your Fortanix API access key. You can obtain this key by logging into your Fortanix account and navigating to the 'API Keys' page in the 'Settings' section." prefix:"kms.list.fortanix." env:"SC_KMS_LIST_FORTANIX_API_KEY" required:""`
}

func initializeKmsListFortanixCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cli CLI

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsListFortanixGroupsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsListFortanixGroupsCmd", e)
		}
		return err
	}
	if cli, err = SecretumFlags(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsListFortanixGroupsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsListFortanixGroupsCmd", e)
		}
		return err
	}
	uri := fmt.Sprintf("keymanager:fortanix?endpoint=%s&apikey=%s", url.QueryEscape(cli.Kms.List.Fortanix.Flags.ApiEndpointURL), url.QueryEscape(cli.Kms.List.Fortanix.Flags.ApiKey))
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithKeyManager(uri, c.Apps.Logger),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsListFortanixGroupsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsListFortanixGroupsCmd", e)
		}
		return err
	}
	*c = common.Cmdctx{
		Cmd:      c.Cmd,
		InitSeq:  c.InitSeq,
		Apps:     c.Apps,
		Adapters: c.Adapters,
		Ports:    c.Ports,
	}

	return nil
}

func (cmd *KmsListFortanixCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsListFortanixCmd}, c.InitSeq...)

	return nil
}
