package list

import (
	"errors"
	"fmt"
	"github.com/bf2fc6cc711aee1a0c2a/cos-tools/rhoc/pkg/util/cmdutil"
	"net/http"
	"strconv"
	"time"

	"github.com/bf2fc6cc711aee1a0c2a/cos-tools/rhoc/pkg/api/admin"
	"github.com/bf2fc6cc711aee1a0c2a/cos-tools/rhoc/pkg/dumper"
	"github.com/bf2fc6cc711aee1a0c2a/cos-tools/rhoc/pkg/service"
	"github.com/olekukonko/tablewriter"
	"github.com/redhat-developer/app-services-cli/pkg/core/cmdutil/flagutil"
	"github.com/redhat-developer/app-services-cli/pkg/core/ioutil/dump"
	"github.com/redhat-developer/app-services-cli/pkg/shared/factory"
	"github.com/spf13/cobra"
)

type connector struct {
	ID              string
	NamespaceID     string
	Owner           string
	CreatedAt       time.Time
	ModifiedAt      time.Time
	ConnectorTypeId string
	Revision        int64
	DesiredState    string
	State           string
}

type connectorWide struct {
	ID              string
	Name            string
	NamespaceID     string
	Owner           string
	CreatedAt       time.Time
	ModifiedAt      time.Time
	ConnectorTypeId string
	Revision        int64
	DesiredState    string
	State           string
	Error           string
}

type options struct {
	outputFormat string
	page         int
	limit        int
	all          bool
	clusterID    string
	orderBy      string
	search       string
	namespaceID  string

	f *factory.Factory
}

func NewListCommand(f *factory.Factory) *cobra.Command {
	opts := options{
		f: f,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list",
		Long:    "list",
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.outputFormat != "" && !flagutil.IsValidInput(opts.outputFormat, cmdutil.ValidOutputs()...) {
				return flagutil.InvalidValueError("output", opts.outputFormat, cmdutil.ValidOutputs()...)
			}
			if opts.clusterID != "" && opts.namespaceID != "" {
				return errors.New("set either cluster-id or namespace-id, not both")
			}
			if opts.clusterID == "" && opts.namespaceID == "" {
				return errors.New("either cluster-id or namespace-id are required")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(&opts)
		},
	}

	cmdutil.AddOutput(cmd, &opts.outputFormat)
	cmdutil.AddPage(cmd, &opts.page)
	cmdutil.AddLimit(cmd, &opts.limit)
	cmdutil.AddAllPages(cmd, &opts.all)
	cmdutil.AddOrderBy(cmd, &opts.orderBy)
	cmdutil.AddSearch(cmd, &opts.search)
	cmdutil.AddClusterID(cmd, &opts.clusterID)
	cmdutil.AddNamespaceID(cmd, &opts.namespaceID)

	return cmd
}

func run(opts *options) error {
	c, err := service.NewAdminClient(&service.Config{
		F: opts.f,
	})
	if err != nil {
		return err
	}

	items := admin.ConnectorAdminViewList{
		Kind:  "ConnectorAdminViewList",
		Items: make([]admin.ConnectorAdminView, 0),
		Total: 0,
		Size:  0,
	}

	for i := opts.page; i == opts.page || opts.all; i++ {
		var result *admin.ConnectorAdminViewList
		var err error
		var httpRes *http.Response

		if opts.clusterID != "" {
			e := c.ConnectorClustersAdminApi.GetClusterConnectors(opts.f.Context, opts.clusterID)
			e = e.Page(strconv.Itoa(i))
			e = e.Size(strconv.Itoa(opts.limit))

			if opts.orderBy != "" {
				e = e.OrderBy(opts.orderBy)
			}
			if opts.search != "" {
				e = e.Search(opts.search)
			}

			result, httpRes, err = e.Execute()
		}

		if opts.namespaceID != "" {
			e := c.ConnectorClustersAdminApi.GetNamespaceConnectors(opts.f.Context, opts.namespaceID)
			e = e.Page(strconv.Itoa(i))
			e = e.Size(strconv.Itoa(opts.limit))

			if opts.orderBy != "" {
				e = e.OrderBy(opts.orderBy)
			}
			if opts.search != "" {
				e = e.Search(opts.search)
			}

			result, httpRes, err = e.Execute()
		}

		if httpRes != nil {
			defer func() {
				_ = httpRes.Body.Close()
			}()
		}
		if err != nil {
			if httpRes != nil && httpRes.StatusCode == http.StatusInternalServerError {
				e, _ := service.ReadError(httpRes)
				if e.Reason != "" {
					err = fmt.Errorf("%s: [%w]", err.Error(), errors.New(e.Reason))
				}
			}
			return err
		}
		if result == nil || len(result.Items) == 0 {
			break
		}

		items.Items = append(items.Items, result.Items...)
		items.Size = int32(len(items.Items))
		items.Total = result.Total
	}

	if len(items.Items) == 0 && opts.outputFormat == "" {
		opts.f.Logger.Info("No result")
		return nil
	}

	switch opts.outputFormat {
	case dump.EmptyFormat:
		dumpAsTable(opts.f, items, false)
	case "wide":
		dumpAsTable(opts.f, items, true)
	default:
		return dump.Formatted(opts.f.IOStreams.Out, opts.outputFormat, items)
	}

	return nil
}

func dumpAsTable(f *factory.Factory, items admin.ConnectorAdminViewList, wide bool) {
	r := make([]interface{}, 0, len(items.Items))

	for i := range items.Items {
		k := items.Items[i]

		if wide {
			r = append(r, connectorWide{
				NamespaceID:     k.NamespaceId,
				ID:              k.Id,
				Name:            k.Name,
				Owner:           k.Owner,
				CreatedAt:       k.CreatedAt,
				ModifiedAt:      k.ModifiedAt,
				ConnectorTypeId: k.ConnectorTypeId,
				Revision:        k.ResourceVersion,
				DesiredState:    string(k.DesiredState),
				State:           string(k.Status.State),
				Error:           k.Status.Error,
			})
		} else {
			r = append(r, connector{
				NamespaceID:     k.NamespaceId,
				ID:              k.Id,
				Owner:           k.Owner,
				CreatedAt:       k.CreatedAt,
				ModifiedAt:      k.ModifiedAt,
				ConnectorTypeId: k.ConnectorTypeId,
				Revision:        k.ResourceVersion,
				DesiredState:    string(k.DesiredState),
				State:           string(k.Status.State),
			})
		}
	}

	t := dumper.NewTable(map[string]func(s string) tablewriter.Colors{
		"state": statusCustomizer,
	})

	t.Dump(r, f.IOStreams.Out)
}

func statusCustomizer(s string) tablewriter.Colors {
	switch s {
	case "ready":
		return tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiGreenColor}
	case "failed":
		return tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiRedColor}
	case "stopped":
		return tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiYellowColor}
	}

	return tablewriter.Colors{}
}