package params

import (
	"errors"
	"time"

	"github.com/bryanfks-dev/logos/pkg/util"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type SyncPullChangesParam struct {
	Since string `param:"since"`
}

func (p SyncPullChangesParam) ParsedSince() (time.Time, error) {
	return util.ParseDatetimeParam(p.Since)
}

func (p SyncPullChangesParam) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(
			&p.Since,
			validation.Required,
			validation.By(func(value any) error {
				sinceStr, ok := value.(string)
				if !ok {
					return errors.New("must be a valid datetime string")
				}

				if _, err := util.ParseDatetimeParam(sinceStr); err != nil {
					return errors.New(
						"must be a valid RFC3339 or YYYY-MM-DD datetime",
					)
				}

				return nil
			}),
		),
	)
}
