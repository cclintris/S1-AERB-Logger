package constant

/*
*************************************************************

	CONSTANT

*************************************************************
*/

// Declaration of units in bytes
const (
	B float64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)
