package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/slack-go/slack"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewBot(
	client *slack.Client,
	channels []string,
	insulter InsultFactory,
) *bot {
	return &bot{
		client:   client,
		channels: channels,
		insulter: insulter,
	}
}

type bot struct {
	ctx      context.Context
	client   *slack.Client
	channels []string
	insulter InsultFactory
}

func getNextDeadline(hour int, minute int) time.Time {
	year, month, day := time.Now().Date()
	loc := time.Now().Location()
	return time.Date(year, month, day, hour, minute, 0, 0, loc)
}

func getDay() time.Time {
	year, month, day := time.Now().Date()
	loc := time.Now().Location()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func (b *bot) runNotifier(hour int, minute int) error {
	nextDeadline := getNextDeadline(hour, minute)
	fmt.Println(nextDeadline)
	for {
		time.Sleep(time.Until(nextDeadline))
		nextDeadline = nextDeadline.Add(24 * time.Hour)
		for _, channelID := range b.channels {
			if err := b.checkChannel(channelID); err != nil {
				return err
			}
		}
	}
}

func (b *bot) checkChannel(channelID string) error {
	usersInChannel, _, err := b.client.GetUsersInConversation(
		&slack.GetUsersInConversationParameters{
			ChannelID: channelID,
		},
	)
	if err != nil {
		return err
	}

	usersPosted := make(map[string]bool, len(usersInChannel))
	for _, u := range usersInChannel {
		usersPosted[u] = false
	}

	today := getDay()
	todayTS := fmt.Sprintf("%d.000000", today.Unix())
	fmt.Println(todayTS)
	resp, err := b.client.GetConversationHistory(
		&slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    todayTS,
		},
	)
	if err != nil {
		return err
	}
	for _, msg := range resp.Messages {
		// only count texts as real messages
		if msg.SubType != "" {
			continue
		}
		usersPosted[msg.User] = true
	}

	for user, posted := range usersPosted {
		if !posted {
			insult := b.insulter.GetInsult()
			if _, _, err := b.client.PostMessage(
				channelID,
				slack.MsgOptionText(fmt.Sprintf("<@%s>\n Где стендап, блядь?\n %s", user, insult), false),
				// slack.MsgOptionAttachments(
				// slack.Attachment{
				// Text: , user, insult),
				// },
				// ),
			); err != nil {
				return err
			}
		}
	}

	return nil
}
