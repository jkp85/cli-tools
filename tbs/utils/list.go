package utils

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ListFlagsSetter interface {
	SetLimit(*string)
	SetOffset(*string)
	SetOrdering(*string)
}

type ListFlags struct {
	Limit, Offset int
	Order         string
}

func (lf *ListFlags) Apply(l ListFlagsSetter) {
	limit := strconv.Itoa(lf.Limit)
	offset := strconv.Itoa(lf.Offset)
	l.SetLimit(&limit)
	l.SetOffset(&offset)
	l.SetOrdering(&lf.Order)
}

func (lf *ListFlags) Set(cmd *cobra.Command) {
	defaultLimit := viper.GetInt("limit")
	cmd.Flags().IntVar(&lf.Limit, "limit", defaultLimit, "Limit list results")
	cmd.Flags().IntVar(&lf.Offset, "offset", 0, "Offset list results")
	cmd.Flags().StringVar(&lf.Order, "order", "", "Output order")
}
