package poll

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

const (
	correctMessage = `What is your favourite colour?
- Red
- Blue
- Green
- Yellow
24h
`
	noDurationMessage = `What is your favourite colour?
- Red
- Blue
`
	notEnoughOptionsMessage = `What is your favourite colour?
- Red
24h
`
	tooManyOptionsMessage = `What is your favourite colour?
- Red
- Blue
- Green
- Yellow
- Orange
24h
`
	invalidDurationMessage = `What is your favourite colour?
- Red
- Blue
24
)
`
	nonDefaultDurationMessage = `What is your favourite colour?
- Red
- Blue
- Green
- Yellow
1h
`
	multilineQuestionMessage = `Multi
line
question
- Red
- Blue
`
	randomMessage = `The current transactions spec is a great fallback for all kinds of transactions, but in a couple months most transactions in frames won't happen with it imo.

Just as frames allowed embeds by implementing a limited, safe scope, there's an opportunity to do the same with transactions, starting with small payments`

	noSpacesOnOptionsMessage = `What is your favourite colour?
-Red
-Blue`

	otherDurationFormatMessage = `What is your favourite colour?
-Red
-Blue
12 hours`
	otherDurationFormat2Message = `What is your favourite colour?
-Red
-Blue
1 hour`
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
	expectedMultilineQuestionPoll = &Poll{
		Question: "Multi line question",
		Options:  []string{"Red", "Blue"},
		Duration: DefaultConfig.DefaultDuration,
	}
	expectedNoSpacesOnOptionsPoll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue"},
		Duration: DefaultConfig.DefaultDuration,
	}
	expectedOtherDurationFormatPoll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue"},
		Duration: time.Hour * 12,
	}
	expectedOtherDurationFormat2Poll = &Poll{
		Question: "What is your favourite colour?",
		Options:  []string{"Red", "Blue"},
		Duration: time.Hour,
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

	multilineQuestionPoll, err := ParseString(multilineQuestionMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(multilineQuestionPoll.Question, qt.Equals, expectedMultilineQuestionPoll.Question)
	c.Assert(multilineQuestionPoll.Options, qt.ContentEquals, expectedMultilineQuestionPoll.Options)
	c.Assert(multilineQuestionPoll.Duration, qt.Equals, expectedMultilineQuestionPoll.Duration)

	noSpacesOnOptionsPoll, err := ParseString(noSpacesOnOptionsMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(noSpacesOnOptionsPoll.Question, qt.Equals, expectedNoSpacesOnOptionsPoll.Question)
	c.Assert(noSpacesOnOptionsPoll.Options, qt.ContentEquals, expectedNoSpacesOnOptionsPoll.Options)
	c.Assert(noSpacesOnOptionsPoll.Duration, qt.Equals, expectedNoSpacesOnOptionsPoll.Duration)

	otherDurationFormatPoll, err := ParseString(otherDurationFormatMessage, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(otherDurationFormatPoll.Question, qt.Equals, expectedOtherDurationFormatPoll.Question)
	c.Assert(otherDurationFormatPoll.Options, qt.ContentEquals, expectedOtherDurationFormatPoll.Options)
	c.Assert(otherDurationFormatPoll.Duration, qt.Equals, expectedOtherDurationFormatPoll.Duration)

	otherDurationFormat2Poll, err := ParseString(otherDurationFormat2Message, DefaultConfig)
	c.Assert(err, qt.IsNil)
	c.Assert(otherDurationFormat2Poll.Question, qt.Equals, expectedOtherDurationFormat2Poll.Question)
	c.Assert(otherDurationFormat2Poll.Options, qt.ContentEquals, expectedOtherDurationFormat2Poll.Options)
	c.Assert(otherDurationFormat2Poll.Duration, qt.Equals, expectedOtherDurationFormat2Poll.Duration)

	_, err = ParseString(notEnoughOptionsMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrMinOptionsNotReached)

	_, err = ParseString(tooManyOptionsMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrMaxOptionsReached)

	_, err = ParseString(invalidDurationMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrParsingDuration)

	_, err = ParseString(randomMessage, DefaultConfig)
	c.Assert(err, qt.ErrorIs, ErrMinOptionsNotReached)
}
