package poll

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

const (
	correctMessage = `!poll
What is your favourite colour?
- Red
- Blue
- Green
- Yellow
24h
`
	noDurationMessage = `!poll
What is your favourite colour?
- Red
- Blue
`
	notEnoughOptionsMessage = `!poll
What is your favourite colour?
- Red
24h
`
	tooManyOptionsMessage = `!poll
What is your favourite colour?
- Red
- Blue
- Green
- Yellow
- Orange
24h
`
	invalidDurationMessage = `!poll
What is your favourite colour?
- Red
- Blue
24
)
`
	nonDefaultDurationMessage = `!poll
What is your favourite colour?
- Red
- Blue
- Green
- Yellow
1h
`
	unrecognisedCommandMessage = `!lalala
What is your favourite colour?
- Red
- Blue
- Green
- Yellow
1h
`
	inlineQuestionMessage = `!poll What is your favourite colour?
- Red
- Blue
`
)

var (
	expectedCorrectPoll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue", "Green", "Yellow"},
		Duration: time.Hour * 24,
	}
	expectedNoDurationPoll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue"},
		Duration: DefaultConfig.DefaultDuration,
	}
	expectedNonDefaultDurationPoll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue", "Green", "Yellow"},
		Duration: time.Hour,
	}
	expectedInlineQuestionPoll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue"},
		Duration: DefaultConfig.DefaultDuration,
	}
)

func TestParseString(t *testing.T) {
	c := qt.New(t)

	correctPoll, err := ParseString(correctMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(correctPoll.Question, qt.Equals, expectedCorrectPoll.Question)
	c.Assert(correctPoll.Options, qt.ContentEquals, expectedCorrectPoll.Options)
	c.Assert(correctPoll.Duration, qt.Equals, expectedCorrectPoll.Duration)

	noDurationPoll, err := ParseString(noDurationMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(noDurationPoll.Question, qt.Equals, expectedNoDurationPoll.Question)
	c.Assert(noDurationPoll.Options, qt.ContentEquals, expectedNoDurationPoll.Options)
	c.Assert(noDurationPoll.Duration, qt.Equals, expectedNoDurationPoll.Duration)

	nonDefaultDurationPoll, err := ParseString(nonDefaultDurationMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(nonDefaultDurationPoll.Question, qt.Equals, expectedNonDefaultDurationPoll.Question)
	c.Assert(nonDefaultDurationPoll.Options, qt.ContentEquals, expectedNonDefaultDurationPoll.Options)
	c.Assert(nonDefaultDurationPoll.Duration, qt.Equals, expectedNonDefaultDurationPoll.Duration)

	inlineQuestionPoll, err := ParseString(inlineQuestionMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(inlineQuestionPoll.Question, qt.Equals, expectedInlineQuestionPoll.Question)
	c.Assert(inlineQuestionPoll.Options, qt.ContentEquals, expectedInlineQuestionPoll.Options)
	c.Assert(inlineQuestionPoll.Duration, qt.Equals, expectedInlineQuestionPoll.Duration)

	_, err = ParseString(notEnoughOptionsMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrMinOptionsNotReached)

	_, err = ParseString(tooManyOptionsMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrMaxOptionsReached)

	_, err = ParseString(invalidDurationMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrParsingDuration)

	_, err = ParseString(unrecognisedCommandMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrUnrecognisedCommand)
}
