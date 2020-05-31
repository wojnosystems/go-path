package go_path

type componentType uint8

const (
	componentTypeInvalid componentType = iota
	componentTypeStruct
	componentTypeMap
	componentTypeArray
)
