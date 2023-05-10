package secretum

import (
	"fmt"
	"net/url"

	"fry.org/cmo/cli/internal/cli/common"
	"fry.org/cmo/cli/internal/infrastructure"
	"github.com/speijnik/go-errortree"
	"github.com/workanator/go-floc/v3"
)

type KmsFortanixCmd struct {
	Flags   KmsFortanixFlags      `embed:""`
	List    KmsFortanixListCmd    `cmd:"" help:"List fortanix objects."`
	Decrypt KmsFortanixDecryptCmd `cmd:"" help:"Decrypt fortanix objects."`
}

type KmsFortanixFlags struct {
	ApiEndpointURL string `help:"The URL for the Fortanix API endpoint. Make sure to include the trailing slash." prefix:"kms.fortanix." env:"SC_KMS_FORTANIX_API_ENDPOINT_URL" default:"https://kms-test.adidas-group.com"`
	ApiKey         string `help:"Your Fortanix API access key. You can obtain this key by logging into your Fortanix account and navigating to the 'API Keys' page in the 'Settings' section." prefix:"kms.fortanix." env:"SC_KMS_FORTANIX_API_KEY" required:""`
}

func initializeKmsFortanixCmd(ctx floc.Context, ctrl floc.Control) error {
	var err, rcerror error
	var c *common.Cmdctx
	var cmd KmsCmd

	if c, err = common.CommonCmdCtx(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixCmd", e)
		}
		return err
	}
	if cmd, err = SecretumKmsCmd(ctx); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixListGroupsCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixListGroupsCmd", e)
		}
		return err
	}
	uri := fmt.Sprintf("keymanager:fortanix?endpoint=%s&apikey=%s", url.QueryEscape(cmd.Fortanix.Flags.ApiEndpointURL), url.QueryEscape(cmd.Fortanix.Flags.ApiKey))
	infraOptions := []infrastructure.AdapterOption{
		infrastructure.WithKeyManager(uri, c.Apps.Logger),
		infrastructure.WithTablePrinter(),
	}
	if err = infrastructure.AdapterWithOptions(&c.Adapters, infraOptions...); err != nil {
		if e := SecretumSetRCErrorTree(ctx, "initializeKmsFortanixCmd", err); e != nil {
			return errortree.Add(rcerror, "initializeKmsFortanixCmd", e)
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

func (cmd *KmsFortanixCmd) Run(cli *CLI, c *common.Cmdctx, rcerror *error) error {

	c.InitSeq = append([]floc.Job{initializeKmsFortanixCmd}, c.InitSeq...)

	return nil
}
