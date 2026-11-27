package validator

import (
	"errors"

	"github.com/bryanfks-dev/logos/pkg/util"
)

func ValidateUUIDParam(value any) error {
	idStr, ok := value.(string)
	if !ok {
		return errors.New("must be a valid UUID string")
	}

	if _, err := util.ParseUUIDParam(idStr); err != nil {
		return errors.New("must be a valid UUID")
	}

	return nil
}
