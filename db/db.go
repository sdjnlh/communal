package db

import (
	"xorm.io/xorm"
)

var DB *xorm.Engine

const (
	STATUS_COMMON_DELETED = int16(0)
	STATUS_COMMON_OK      = int16(1)

	DEFAULT_PAGE_SIZE = 20

	UndeletedCause = "dtd=false"
)
