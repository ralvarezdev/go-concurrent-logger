package go_concurrent_logger

import (
	"fmt"
	"time"
)

type (
	// Message is the struct to handle log messages
	Message struct {
		Category      Category
		Content       string
		Tag           string
		formattedTime string
	}
)

// NewMessage creates a new Message instance.
//
// Parameters:
//
// category: Category of the log message.
// content: Content of the log message.
// tag: Tag for the log message.
// timestampFormat: Format for the timestamp.
//
// Returns:
//
// A pointer to a Message instance.
func NewMessage(
	category Category,
	content string,
	tag string,
	timestampFormat string,
) *Message {
	return &Message{
		category,
		content,
		tag,
		time.Now().Format(timestampFormat),
	}
}

// String returns the representation of the log message.
//
// Returns:
//
// The formatted log message
func (m *Message) String() string {
	return fmt.Sprintf(
		"%s %s [%s] %s",
		m.formattedTime,
		m.Category.String(),
		m.Tag,
		m.Content,
	)
}
