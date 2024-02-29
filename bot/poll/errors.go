package poll

import "fmt"

var (
	// ErrQuestionNotSet is returned when the poll command is recognised but the
	// question content is not set.
	ErrQuestionNotSet = fmt.Errorf("question content not set")
	// ErrParsingDuration is returned when the poll command is recognised, the
	// question and options are set, and the duration content is set but it
	// cannot be parsed.
	ErrParsingDuration = fmt.Errorf("error parsing duration")
	// ErrMinOptionsNotReached is returned when the poll command is recognised,
	// the question content is set, and the number of options is less than the
	// minimum number of options.
	ErrMinOptionsNotReached = fmt.Errorf("min number of options not reached")
	// ErrMaxOptionsReached is returned when the poll command is recognised, the
	// question content is set, and the number of options is greater than the
	// maximum number of options.
	ErrMaxOptionsReached = fmt.Errorf("max number of options reached")
)
