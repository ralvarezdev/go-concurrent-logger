package go_concurrent_logger

type (
	// Category is an enum to define the category of log message.
	Category uint8
)

const (
	CategoryNil Category = iota
	CategoryInfo
	CategoryWarning
	CategoryError
	CategoryDebug
)

var (
	// CategoryNames maps a given Category to its string name
	CategoryNames = map[Category]string{
		CategoryInfo:    "INFO",
		CategoryWarning: "WARNING",
		CategoryError:   "ERROR",
		CategoryDebug:   "DEBUG",
	}
)

// String returns the string representation of the Category
//
// Returns:
//
// The string representation of the Category enum
func (c Category) String() string {
	return CategoryNames[c]
}
