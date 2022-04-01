package proto

type ServerPropertyFields string

const (
	Type     ServerPropertyFields = "virt_type"
	Location ServerPropertyFields = "location"
	Region   ServerPropertyFields = "region"
)

var DefaultPropertyLabelMapping map[ServerPropertyFields]string

type TimeRangeType string

const (
	THIS_MONTH_SO_FAR TimeRangeType = "THIS_MONTH_SO_FAR"
)

func init() {
	DefaultPropertyLabelMapping = make(map[ServerPropertyFields]string)
	DefaultPropertyLabelMapping[Type] = string(Type)
	DefaultPropertyLabelMapping[Location] = string(Location)
	DefaultPropertyLabelMapping[Region] = string(Region)
}
