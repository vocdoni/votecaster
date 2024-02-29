// poll package contains the logic to parse a poll message from a string and
// create a poll struct with the question, options and duration. The package
// also contains the default configuration for a poll.
// The parser follows the format:
// !poll <question>
// - <option 1>
// - <option 2>
// - <option 3*>
// - <option 4*>
// <duration*>
// The duration is optional and if not set, it takes the default duration. The
// minimum and maximum number of options are also configurable. The question
// can be set in multiple lines, but the options must be set in a single line.
package poll

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	optionPrefix = "-"
	linebreak    = "\n"
	space        = " "
)

// DefaultConfig var contains the default configuration for a poll with a
// minimum of 2 options, a maximum of 4 options, a minimum duration of 1 hour,
// a maximum duration of 15 days and a default duration of 24 hours.
var DefaultConfig = &PollConfig{
	MinOptions:      2,
	MaxOptions:      4,
	MinDuration:     time.Hour,
	MaxDuration:     24 * time.Hour * 15, // 15 days
	DefaultDuration: time.Hour * 24,
}

// PollConfig struct contains the configuration for a poll with a minimum and
// maximum number of options and a default and maximum duration.
type PollConfig struct {
	MinOptions      int
	MaxOptions      int
	MinDuration     time.Duration
	MaxDuration     time.Duration
	DefaultDuration time.Duration
}

// Poll represents a poll with a question, options and duration
type Poll struct {
	Question string
	Options  []string
	Duration time.Duration
}

// ParseString parses a string message and returns a Poll struct with the
// question, options and duration. The message should follow the format:
// !poll <question>
// - <option 1>
// - <option 2>
// - <option 3*>
// - <option 4*>
// <duration*>
// The duration is optional and by default is 24 hours. If the message does not
// follow the format, an error is returned.
func ParseString(message string, config *PollConfig) (*Poll, error) {
	// create vars to store the question, options and duration
	var question string
	var options []string
	var duration time.Duration = config.DefaultDuration
	// poll message follows the format:
	// !poll <question>
	// - <option 1>
	// - <option 2>
	// - <option 3*>
	// - <option 4*>
	// <duration*>

	// create a new reader from the message content and a new scanner from
	// the reader
	reader := strings.NewReader(message)
	scanner := bufio.NewScanner(reader)
	// read every line from the message
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// if the line is empty, continue
		if line == "" {
			continue
		}
		// line is a <question> if:
		//  - it not starts with a dash
		//  - any question has been set
		// line is a <option-n> if:
		//  - it starts with a dash
		//  - the question has been set
		//  - the number of options is less than the max number of options
		// line is a <duration> if:
		//  - it not starts with a dash
		//  - the question has been set
		//  - at least the min number of options has been set
		//  - by default, the duration is 24 hours
		startWithDash := strings.HasPrefix(line, optionPrefix)
		numOfQuestions := len(options)
		if !startWithDash {
			// if the line is a question append it to the question and continue
			if numOfQuestions == 0 {
				question += fmt.Sprintf("%s%s", line, linebreak)
				continue
			}
			// if the line is a duration, try to parse it, if it fails, return
			// an error, otherwise, break the loop and return the result
			var err error
			if duration, err = time.ParseDuration(line); err != nil {
				return nil, errors.Join(ErrParsingDuration, err)
			}
			if duration < config.MinDuration || duration > config.MaxDuration {
				return nil, fmt.Errorf("duration out of range: %w", ErrParsingDuration)
			}
			break
		}
		// if the line is an option and the number of options is greater than
		// the max number of options, return an error
		if numOfQuestions >= config.MaxOptions {
			return nil, fmt.Errorf("%w: %d", ErrMaxOptionsReached, config.MaxOptions)
		}
		// append the option to the options
		optionText := strings.TrimSpace(strings.TrimPrefix(line, optionPrefix))
		options = append(options, optionText)
	}
	// check poll content
	if question == "" {
		return nil, ErrQuestionNotSet
	}
	if len(options) < config.MinOptions {
		return nil, fmt.Errorf("%w: %d", ErrMinOptionsNotReached, config.MinOptions)
	}
	// return the results
	return &Poll{
		Question: strings.ReplaceAll(question, linebreak, space),
		Options:  options,
		Duration: duration,
	}, nil
}
