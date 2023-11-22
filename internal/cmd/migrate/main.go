package migrate

import (
	"io"
	"log"
	"os"

	_ "ariga.io/atlas-go-sdk/recordriver" // required by atlasgo
	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func main(shutdonwer fx.Shutdowner) {
	stmts, err := gormschema.New("postgres").Load(new(url.URL))
	if err != nil {
		log.Fatalf("failed to load gorm schema %s", err)
	}

	_, _ = io.WriteString(os.Stdout, stmts)

	_ = shutdonwer.Shutdown()
}

// Register migrate command.
func Register(root *cobra.Command) {
	root.AddCommand(
		//nolint: exhaustruct
		&cobra.Command{
			Use:   "migrate",
			Short: "Database migration",
			Run: func(_ *cobra.Command, _ []string) {
				fx.New(
					fx.NopLogger,
					fx.Invoke(main),
				).Run()
			},
		},
	)
}
