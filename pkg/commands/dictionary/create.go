package dictionary

import (
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// CreateCommand calls the Fastly API to create a service.
type CreateCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.CreateDictionaryInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	writeOnly cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("create", "Create a Fastly edge dictionary on a Fastly service version")
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Name of Dictionary").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("write-only", "Whether to mark this dictionary as write-only. Can be true or false (defaults to false)").Action(c.writeOnly.Set).StringVar(&c.writeOnly.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	if c.writeOnly.WasSet {
		writeOnly, err := strconv.ParseBool(c.writeOnly.Value)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Service ID":      serviceID,
				"Service Version": serviceVersion.Number,
			})
			return err
		}
		c.Input.WriteOnly = fastly.Compatibool(writeOnly)
	}

	d, err := c.Globals.APIClient.CreateDictionary(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	var writeOnlyOutput string
	if d.WriteOnly {
		writeOnlyOutput = "as write-only "
	}

	text.Success(out, "Created dictionary %s %s(service %s version %d)", d.Name, writeOnlyOutput, d.ServiceID, d.ServiceVersion)
	return nil
}
