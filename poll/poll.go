package poll

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	optionPrefix    = "-"
	lineBreakSuffix = "\n"
	pollCommand     = "!poll"
)

var DefaultConfig = PollConfig{
	MinOptions:      2,
	MaxOptions:      4,
	MinDuration:     time.Hour,
	MaxDuration:     time.Hour * 8760, // 1 year
	DefaultDuration: time.Hour * 24,
}

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
// !poll
// <question>
// - <option 1>
// - <option 2>
// - <option 3*>
// - <option 4*>
// <duration*>
// The duration is optional and by default is 24 hours. If the message does not
// follow the format, an error is returned.
func ParseString(message string, config PollConfig) (*Poll, error) {
	// create a flag to check if the command has been recognised
	recognisedCommand := false
	// create vars to store the question, options and duration
	var question string
	var options []string
	var duration time.Duration = config.DefaultDuration
	// poll message follows the format:
	// !poll
	// <question>
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
		// if the line contains the command, set the flag and continue
		if strings.HasPrefix(line, pollCommand) {
			recognisedCommand = true
			question = strings.TrimSpace(strings.TrimPrefix(line, pollCommand))
			continue
		}
		// if the line is not a command, and the command has not been
		// recognised, return an error
		if !recognisedCommand {
			return nil, ErrUnrecognisedCommand
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
				question += fmt.Sprintf("%s%s", line, lineBreakSuffix)
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
		Question: strings.TrimSuffix(question, lineBreakSuffix),
		Options:  options,
		Duration: duration,
	}, nil
}
