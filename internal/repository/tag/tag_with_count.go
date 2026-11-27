package tag

import "github.com/bryanfks-dev/logos/pkg/entity"

type tagWithCount struct {
	entity.Tag
	UsageCount int `gorm:"column:usage_count"`
}
