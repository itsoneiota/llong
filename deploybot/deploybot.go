package deploybot

import (
	"fmt"
	"regexp"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
	"github.com/owenmorgan/llongdocker"
)

// DeployBot represents a bot for deployment.
type DeployBot struct {
	DockerClient *llongdocker.LlongDockerClient
}

func init() {
	slick.RegisterPlugin(&DeployBot{})
}

// InitPlugin initialises the plugin.
func (dep *DeployBot) InitPlugin(bot *slick.Bot) {
	bot.Listen(&slick.Listener{
		MessageHandlerFunc: dep.ChatHandler,
	})
	dep.DockerClient = llongdocker.NewLlongDockerClient(
		"eu-west-1",
		"unix:///var/run/docker.sock",
		"https://597304777786.dkr.ecr.eu-west-1.amazonaws.com",
		"AWS",
		"AQECAHhwm0YaISJeRtJm5n1G6uqeekXuoXXPe5UFce9Rq8/14wAAAvEwggLtBgkqhkiG9w0BBwagggLeMIIC2gIBADCCAtMGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMfNrkvy7IJJH5bqgTAgEQgIICpCQl5S5WWsi3S0k1HbYZKZkJmYJBVQJfFO7sm49wMdBV3JKqAUocFABloOgRgRINhvm3WUPRPrP+vjAySeKDoRMeBoVuyVi+p6N2MfRtTZa+mHt4oPFL3lZNJjADhr/LuAdMp/hKPs7vOjiXWg2k6KgDApbCa7qb+gvOGWUtkIrIkwJ23Y5i90xYbex7jVjjj3C3iL9d0tmI8SCEwL4/mDL1Fd0RAQHpisvR0vnKJHx8u7rKQtpsJirqfgmp5EOwC17MQjc+thdotwHlOanxsonh9FtBSd463YQvQj6f+pMlLu5uJjurf8eOnsRORCIk94AfHZjlHbHWmhldcXO2WWvcmFQniGAjo3jAeTYuI2F5ex2ZNqBfNWQBXfzebd6TEvnfStDaCzg5uVFhJ7xTtEk8p4AoNnGTzpOuqTJL0TE1uIlyi5HSgbhqzaJ1Sq4AF9nh19Dph3QnYnTIHvVSvw9FB6Mpb3P2Sbd7xDgaG70+QXLWg9KAitlLh56xAVtMuIEaYPv7r9Rxj0m4Hva1OvVULwRLe8MbfxPno1hoErKtG6Gbdpi9Sl0RZUYb4tJH7qEqGFi2ev4/1LSygpQpXS8JLH6h67wzrM8NugxnnJRJiJhHDvjrvg3YXrzk7CxuQGFtogyy8dpxQ6wKTsoz9LR674qygXFRtoII4BNLSe9WlQo7XmdpY66cIIFeHFsECAc+AtiFYBFmu31o2ynBtwesaoY8hDRXiwZZponig3P3RCqSmEQEarEyzzAzH/UVAx5iwuvqDkobXTGrI7PFkKIHoXRGZT/Q7jZusxh3V3MTQGymkov6FNZ36QaW2oZafhEjZwri9Va+0JQLA2vtwr9Wd4w7Hcgf4tQQZUB5q6rz7ATyfklTo9LHmJxLmElkj5ffKms=",
	)
}

var deployFormat = regexp.MustCompile(`(?:deploy|ship)$`)
var deployFromFormat = regexp.MustCompile(`(?:deploy|ship) ([a-zA-Z0-9_\.-]+)$`)
var deployFromToFormat = regexp.MustCompile(`(?:deploy|ship) ([a-zA-Z0-9_\.-]+) to ([a-z_-]+)$`)
var listImagesFormat = regexp.MustCompile(`list (images for )?([a-zA-Z_-]+)`)

// ChatHandler handles chat events.
func (dep *DeployBot) ChatHandler(listen *slick.Listener, msg *slick.Message) {

	if msg.Contains("beer") {
		go func() {
			msg.AddReaction("beer")
		}()
	}

	if msg.MentionsMe {
		if msg.ContainsAny([]string{"thanks", "thank you", "thx", "thnks", "cheers"}) {
			msg.Reply("My pleasure.")
		}
	}

	// Serious stuff now.

	// Discard non "mention_name, " prefixed messages
	// if !strings.HasPrefix(msg.Text, fmt.Sprintf("%s, ", bot.Config.Nickname)) {
	// 	return
	// }

	var match []string
	if match = deployFromToFormat.FindStringSubmatch(msg.Text); match != nil {
		dep.confirmDeploy(listen.Bot, msg, match[1], match[2])
		// reaction = "100"
	} else if match = listImagesFormat.FindStringSubmatch(msg.Text); match != nil {
		dep.listImages(listen.Bot, msg, match[2])
	} else {
		msg.Reply("Sorry, I'm not sure what you mean there.")
	}
}

func (dep *DeployBot) listImages(bot *slick.Bot, msg *slick.Message, repo string) {
	msg.Reply("Listing images for %s…", repo)
	imageHistory, _ := dep.DockerClient.GetRepoImages(repo)

	actions := []slack.AttachmentAction{}
	for _, img := range imageHistory.ImageIds {
		fmt.Println(img)
		displayName := *img.ImageDigest
		if img.ImageTag != nil {
			displayName = *img.ImageTag
		}
		value := "https://597304777786.dkr.ecr.eu-west-1.amazonaws.com/" + repo
		button := slack.AttachmentAction{
			Name:  value,
			Text:  displayName,
			Type:  "button",
			Value: value,
		}
		actions = append(actions, button)
	}

	params := slack.NewPostMessageParameters()
	params.AsUser = true
	params.Username = "llong"
	attachment := slack.Attachment{
		Fallback:   "No joy",
		CallbackID: "123",
		Color:      "#FF0000",
		Actions:    actions,
	}
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := bot.Slack.PostMessage(msg.Channel, "Which image would you like me to ship?", params)
	if err != nil {
		fmt.Printf("NO FUCKING BUTTONS %s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

// Show yes/no buttons to confirm the deploy.
func (dep *DeployBot) confirmDeploy(bot *slick.Bot, msg *slick.Message, from, to string) {
	params := slack.NewPostMessageParameters()
	params.AsUser = true
	params.Username = "llong"
	attachment := slack.Attachment{
		// Pretext: "some pretext",
		Text: "Are you sure?",
		// Uncomment the following part to send a field too
		Fallback:   "No joy",
		CallbackID: "foobar",
		Color:      "#FF0000",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "YES",
				Text:  "Make it so.",
				Type:  "button",
				Value: "YES",
			},
			slack.AttachmentAction{
				Name:  "NO",
				Text:  "Woah",
				Type:  "button",
				Value: "NO",
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := bot.Slack.PostMessage(msg.Channel, fmt.Sprintf("Ship %s to %s?", from, to), params)
	if err != nil {
		fmt.Printf("NO FUCKING BUTTONS %s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
