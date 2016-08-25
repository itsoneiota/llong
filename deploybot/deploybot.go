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
		"AQECAHh+dS+BlNu0NxnXwowbILs115yjd+LNAZhBLZsunOxk3AAAAvEwggLtBgkqhkiG9w0BBwagggLeMIIC2gIBADCCAtMGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMd90K7Qe01TJN2HZwAgEQgIICpIcbw6UQlHWwwxB/ibJV2t+rAuCfWLL41sRKdfjrpHTC+/YVkMA3679bCqsg2XPE73i6offn5DLW5/KjljySpFvq0jK/8n0su9xSki1NV4D6OBhpBCY4wMQXc+qOXKVj1/5BYJoo1TDE8hxVf946fkAQctXfUEN9vdnxreknPGaNC1UlIlLGZ2xhPiZ0t7cjiXzk1wpwsocf39SSHHF7LYbrEkMFhEXBsTqCP7zx2uo9AFonYMFA6fqmWELyansPVvXytRHF3g7iMY/8++uWLMWdsq9I/Waproosphzl4fUSmOobzomp09x1eOfZvdyI4WRKGFFsHG4T4mWB1vSbJIDVl0lr64LcwjauE0IlcXgtdPvv5lMAjz+tqdZ6MnqQkZMObXLx8xnzyHRjnTx3SwmBTfHatFVeK9RC3pW9DTIYfStVTgZOdbDgZeuadkWD8fJqW+hH3XtCkOnBnDaeyOTnWnnS+ANWfLaUy2tWFRmDnlcOOkB2riNA70hYkmj+aXRJF52pi6HEDMUqp+zhCSGfew82FSKc9nFDAX2pyod1M9gxuwJC6FReyiimyHH7jT+qiVSvntPMBABWZWuZBbH7/iIZKPIsiHth//7bS71pMwim5HjJLPB+WFHuC1/xEbwkQmzmbpJWFmWDyST6H6VuCHUxes/mZKoSycWTtDqPwTyyU54xvSiIX/zT3XI8AXxdtDp5afbkADhwy9950DO7P1K5XxBKGPYzB18TLu7dytHBDp8HCt5YaxAVTNqR7VJZjUGyt3uXCCunJMO/Ktkq8xN53u+L8X3JiBsh+G90DqS7A5DEEcpTlkztT/OYCERYVblLFZYOFpBARSaea4NXHLjUT5wHlniAZ4AO76u6M4hV4bNJYDydyKbM6isIazXuYIE=",
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
		value := *img.ImageDigest
		button := slack.AttachmentAction{
			Name:  value,
			Text:  displayName,
			Type:  "button",
			Value: value,
		}
		actions = append(actions, button)
	}

	params := slack.PostMessageParameters{}
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
	params := slack.PostMessageParameters{}
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
	channelID, timestamp, err := bot.PubRTM.PostMessage(msg.Channel, fmt.Sprintf("Ship %s to %s?", from, to), params)
	if err != nil {
		fmt.Printf("NO FUCKING BUTTONS %s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
