package deploybot

import (
	"fmt"
	"regexp"

	"github.com/abourget/slick"
	"github.com/nlopes/slack"
)

// DeployBot represents a bot for deployment.
type DeployBot struct {
}

func init() {
	slick.RegisterPlugin(&DeployBot{})
}

// InitPlugin initialises the plugin.
func (dep *DeployBot) InitPlugin(bot *slick.Bot) {
	bot.Listen(&slick.Listener{
		MessageHandlerFunc: dep.ChatHandler,
	})
}

var deployFormat = regexp.MustCompile(`deploy$`)
var deployFromFormat = regexp.MustCompile(`deploy ([a-zA-Z0-9_\.-]+)$`)
var deployFromToFormat = regexp.MustCompile(`deploy ([a-zA-Z0-9_\.-]+) to ([a-z_-]+)$`)
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

	var reaction string
	var match []string
	if match = deployFormat.FindStringSubmatch(msg.Text); match != nil {
		reaction = "beer"
	} else if match = deployFromFormat.FindStringSubmatch(msg.Text); match != nil {
		reaction = "thumbsup"
	} else if match = deployFromToFormat.FindStringSubmatch(msg.Text); match != nil {
		reaction = "100"
	} else if match = listImagesFormat.FindStringSubmatch(msg.Text); match != nil {
		dep.listImages(msg, match[2])
	} else {
		msg.Reply("Sorry, I'm not sure what you mean there.")
	}
	if reaction != "" {
		msg.AddReaction(reaction)
	}
}

func (dep *DeployBot) listImages(msg *slick.Message, repo string) {
	msg.Reply("Listing images for %s…", repo)
}

// Show yes/no buttons to confirm the deploy.
func (dep *DeployBot) confirmDeploy(bot *slick.Bot, msg *slick.Message) {
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Pretext: "some pretext",
		Text:    "some text",
		// Uncomment the following part to send a field too
		Fallback:   "No joy",
		CallbackID: "123",
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "YES",
				Text:  "Yes",
				Type:  "button",
				Value: "YES",
			},
			slack.AttachmentAction{
				Name:  "NO",
				Text:  "No",
				Type:  "button",
				Value: "NO",
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := bot.Slack.PostMessage(msg.Channel, "Some text", params)
	if err != nil {
		fmt.Printf("NO FUCKING BUTTONS %s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
