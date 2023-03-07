package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func handleEvent(event slackevents.EventsAPIEvent, client *slack.Client, userPost *sync.Map) error {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			err := handleAppMentionEvent(ev, client, userPost)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Unsupporte event type")
	}
	return nil
}

func handleAppMentionEvent(event *slackevents.AppMentionEvent, client *slack.Client, userPost *sync.Map) error {
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}
	timeStamp, _ := strconv.ParseFloat(event.TimeStamp, 8)
	userPost.Store(user.ID, int(timeStamp))
	return nil
}

func swear(user string, channel string, messages []string, client *slack.Client) {
	attachment := slack.Attachment{}
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(messages))
	attachment.Text = "<@" + user + "> " + messages[index]
	client.PostMessage(channel, slack.MsgOptionAttachments(attachment))
}

func mainnn() {
	godotenv.Load(".env")

	messages := []string{}

	file, _ := os.Open("./messages")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		messages = append(messages, scanner.Text())
	}

	data := os.Getenv("USERS")
	users := strings.Split(string(data), ",")
	var userPost sync.Map

	channel := os.Getenv("SLACK_CHANNEL_ID")
	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				switch event.Type {
				case socketmode.EventTypeEventsAPI:
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					socketClient.Ack(*event.Request)
					err := handleEvent(eventsAPIEvent, client, &userPost)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}(ctx, client, socketClient)

	go func() {
		today := time.Now().Format("2006-01-02")
		deadlineTime, _ := time.Parse("2006-01-02 3:04:05 PM", today+" 9:10:00 AM")
		for _, u := range users {
			userPost.Store(u, int(deadlineTime.Add(10*time.Second).Unix()))
		}
		for {
			currentTime := time.Now()
			if currentTime.After(deadlineTime) {
				userPost.Range(func(user, postTime interface{}) bool {
					if time.Unix(int64(postTime.(int)), 0).After(deadlineTime) {
						swear(user.(string), channel, messages, client)
					}
					userPost.Store(user.(string), deadlineTime.Add(24*time.Hour+10*time.Second))
					return true
				})
				deadlineTime = deadlineTime.Add(24 * time.Hour)
			}
		}
	}()

	socketClient.Run()
}

func main() {
	token := "xoxb-4842928093063-4858049813043-BF1lNzC6OEegRzXSVAMnyXtY"
	client := slack.New(token, slack.OptionDebug(false))

	insuleter, err := NewInsultFactory("messages")
	if err != nil {
		panic(err)
	}

	bot := NewBot(
		client,
		[]string{"C04R79YQ8LB"},
		insuleter,
	)
	if err := bot.runNotifier(3, 13); err != nil {
		panic(err)
	}
}
