package poll

import "fmt"

var (
	ErrUnrecognisedCommand  = fmt.Errorf("unrecognised command")
	ErrQuestionNotSet       = fmt.Errorf("question content not set")
	ErrParsingDuration      = fmt.Errorf("error parsing duration")
	ErrMinOptionsNotReached = fmt.Errorf("min number of options not reached")
	ErrMaxOptionsReached    = fmt.Errorf("max number of options reached")
)
