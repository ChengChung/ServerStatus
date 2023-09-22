package proto

type ServerPropertyFields string

const (
	Type     ServerPropertyFields = "virt_type"
	Location ServerPropertyFields = "location"
	Region   ServerPropertyFields = "region"
)

var DefaultPropertyLabelMapping map[ServerPropertyFields]string = map[ServerPropertyFields]string{
	Type:     string(Type),
	Location: string(Location),
	Region:   string(Region),
}

type TimeRangeType string

const (
	THIS_MONTH_SO_FAR TimeRangeType = "THIS_MONTH_SO_FAR"
)
