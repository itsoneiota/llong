package deploybot

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"

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
		"AQECAHh+dS+BlNu0NxnXwowbILs115yjd+LNAZhBLZsunOxk3AAAAvEwggLtBgkqhkiG9w0BBwagggLeMIIC2gIBADCCAtMGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQM12TihiUHivwxcD1NAgEQgIICpAHNjLeajdFspt/ajWKN4CW+PJLBRtO3kP1MaMRsqmWv7V1I81frYy7pJnsfl3lcCnbidGUqvEOTx9pK/kE5wFEQterZJoaB805np5uC1xIkMV67UA6Kpd3ZUEgNvFGjBkmV5u15CJVlSawyLV3W1xxy31eJkKkbgZM+GhG/FqarapqhKJTV/21dZprVSa+f4RSMgJ6zXKfU1q3tdg1sazSaU6jmNsV9Pw9lbkin1P84KalTu4e5SGSzIfbqmGYQ4IYSKYjhjaANgNl8JCbuTuBcOKoPMhoEPI4b0SDrWBWEtPAYcAhbd9LIg9RryPIV2mRKkRm1fNesnSzohKu7MhKvJJyUrPMJpAsJphobdgmXmzV0XpITpYNgV0DCLKrJqiyzZK+ZhyI0UdvlZccq0PxtteltwDIdn2ISwnZwXsyMjtn+wH3NXHGVMtqj3GYhdFxATK3Y6iYOu6uHuHUlRuAgCRgsumyCfTU9EXoyBQ3H02y38D6tMH7k8/yHYXOp4MLna20RqGIw7hV6+FwlTkbp8TWJzQNfG5C7eQPd0Dgb6btigufZgnmePlwVdccy+CeS6obo0pSEAvE4t4q7oKAH1kTvb1df/ygGSFsWEf18f68le8tNpafPVIYS9ZZ0VBsd9SPtuPhjXMkWWF4OepJ/rgnOmBWje3DXV03SwIsEJIhCR5BPL/DrTFqXG7dFxCviOT3krnM3tIvB1cUMM9W6f3JkrspJd3bbBSMrbnWkMKL/mi4uX1ZwxwLtl9v096vcTjPZgEV/AXGIMkc26Wo4Il/gy334+TJ7bqBQb4dENoC0HpCXmFM8cvzHp7/ETm6U8NudZrDsVnLfGo9DJW5fVAByeO03T/Oq4R2Yp9VGhw6LftRUqPDk/31zmfs+skk3nrE=",
	)
}

var cy = false

// var deployFormat = regexp.MustCompile(`(?:deploy|ship)$`)
// var deployFromFormat = regexp.MustCompile(`(?:deploy|ship) ([a-zA-Z0-9_\.-]+)$`)
var deployFromToFormat = regexp.MustCompile(`(?:deploy|ship) ([a-zA-Z0-9_\.-]+) to ([a-z_-]+)$`)
var listImagesFormat = regexp.MustCompile(`list (?:images for )?([a-zA-Z_-]+)`)

var deployFromToFormatCy = regexp.MustCompile(`cludwch ([a-zA-Z0-9_\.-]+) at ([a-z_-]+)$`)
var listImagesFormatCy = regexp.MustCompile(`rhestrwch (?:delwau am )?([a-zA-Z_-]+)`)

// ChatHandler handles chat events.
func (dep *DeployBot) ChatHandler(listen *slick.Listener, msg *slick.Message) {
	cy = false

	if msg.Contains("beer") {
		go func() {
			msg.AddReaction("beer")
		}()
	}

	if msg.MentionsMe {
		if msg.ContainsAny([]string{"thanks", "thank you", "thx", "thnks", "cheers"}) {
			msg.Reply("My pleasure.")
			return
		}
		if msg.ContainsAny([]string{"diolch"}) {
			msg.Reply("Â chroeso.")
			return
		}
	}

	// Serious stuff now.

	var match []string
	if match = deployFromToFormat.FindStringSubmatch(msg.Text); match != nil {
		dep.deploy(msg, match[1], match[2])
		msg.AddReaction("thumbsup")
	} else if match = deployFromToFormatCy.FindStringSubmatch(msg.Text); match != nil {
		cy = true
		dep.deploy(msg, match[1], match[2])
		msg.AddReaction("thumbsup")
		return
	} else if match = listImagesFormat.FindStringSubmatch(msg.Text); match != nil {
		fmt.Printf("%v", match)
		dep.listImages(listen.Bot, msg, match[1])
	} else if match = listImagesFormatCy.FindStringSubmatch(msg.Text); match != nil {
		cy = true
		dep.listImages(listen.Bot, msg, match[1])
	} else {
		msg.Reply("Sorry, I'm not sure what you mean there.")
	}
}

func (dep *DeployBot) listImages(bot *slick.Bot, msg *slick.Message, repo string) {

	if cy {
		msg.Reply("Dw i'n rhestru delwau am %s…", repo)
	} else {
		msg.Reply("Listing images for %s…", repo)
	}
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
	txt := ""
	if cy {
		txt = "Pa ddelw hoffech chi cludo?"
	} else {
		txt = "Which image would you like me to ship?"
	}
	channelID, timestamp, err := bot.Slack.PostMessage(msg.Channel, txt, params)
	if err != nil {
		fmt.Printf("NO FUCKING BUTTONS %s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

// Show yes/no buttons to confirm the deploy.
func (dep *DeployBot) deploy(msg *slick.Message, from, to string) {
	if cy {
		msg.Reply("Dw i'n gweitio arno fe.")
	} else {
		msg.Reply("I'm on it.")
	}
	img := "597304777786.dkr.ecr.eu-west-1.amazonaws.com/" + from
	tag := "dd71ecd"
	err := deploy(dep.DockerClient, img, tag, to)
	if err == nil {
		if cy {
			msg.Reply("Dyna chi.")
		} else {
			msg.Reply("Ta Da!")
		}
	} else {
		if cy {
			msg.Reply(fmt.Sprintf("Roedd problem gyda'ch lleoliad: ```%v```", err))
		} else {
			msg.Reply(fmt.Sprintf("There was a problem with your deployment: ```%v```", err))
		}

	}
}

//Deploy deploys the app
func deploy(client *llongdocker.LlongDockerClient, image, tag, env string) error {

	imgConfig, err := client.GetImageConfig(image, tag)
	if err != nil {
		fmt.Printf("Llongdocker error: %v", err)
		return err
	}
	fmt.Printf("%v", imgConfig)
	imgConfig.Image = image
	imgConfig.Env = env

	memcached := false
	if val, ok := imgConfig.Dependencies["memcached"]; ok {
		memcached = val
	}

	data := struct {
		AppName        string
		AppDescription string
		HostPort       int
		ContainerPort  int
		Env            string
		Image          string

		Memcached bool
	}{
		imgConfig.AppName,
		imgConfig.AppDescription,
		imgConfig.HostPort,
		imgConfig.ContainerPort,
		imgConfig.Env,
		imgConfig.Image,
		memcached,
	}

	t := template.New("app")
	t, err = t.Parse(`
    provider "aws"{
        region = "eu-west-1"
    }

	data "terraform_remote_state" "state" {
		backend = "s3"
		config {
			bucket = "llong"
			key = "{{.AppName}}-{{.Env}}/terraform.tfstate"
			region = "eu-west-1"
		}
	}

    {{if .Memcached}}
    resource "aws_elasticache_cluster" "{{.AppName}}-{{.Env}}-cache" {
        cluster_id = "{{.AppName}}-{{.Env}}"
        engine = "memcached"
        node_type = "cache.t2.micro"
        port = 11211
        num_cache_nodes = 1
        parameter_group_name = "default.memcached1.4"
		subnet_group_name = "${aws_elasticache_subnet_group.{{.AppName}}-{{.Env}}.name}"
    }

	resource "aws_elasticache_subnet_group" "{{.AppName}}-{{.Env}}" {
        name = "{{.AppName}}-{{.Env}}-subnet-group"
        subnet_ids = ["subnet-6373d03b"]
    }
    {{end}}

    resource "aws_ecs_service" "{{.AppName}}-{{.Env}}" {
        name = "{{.AppName}}-{{.Env}}"
        cluster = "arn:aws:ecs:eu-west-1:597304777786:cluster/llong"
        task_definition = "${aws_ecs_task_definition.{{.AppName}}-{{.Env}}.arn}"
        desired_count = 1
    }

       resource "aws_ecs_task_definition" "{{.AppName}}-{{.Env}}" {
        family = "{{.AppName}}-{{.Env}}"
        container_definitions = <<EOF
        [
        {
            "name": "{{.AppName}}-{{.Env}}",
            "image": "{{.Image}}",
            "cpu": 10,
            "memory": 100,
            "essential": true,

            {{if .Memcached}}
             "environment" : [
                { "name" : "memcached_host", "value" : "${aws_elasticache_cluster.{{.AppName}}-{{.Env}}-cache.cache_nodes.0.address}" }
            ],
            {{end}}

            "portMappings": [
            {
                "containerPort": {{.ContainerPort}},
                "hostPort": {{.HostPort}}
            }
            ]
        }
        ]
    EOF
    }`)

	f, err := os.Create("app.tf")
	if err != nil {
		return fmt.Errorf("error creating terraform file: %s", err.Error())
	}

	t.Execute(f, data)

	err = exec.Command("./tfs3", fmt.Sprintf("%s-%s", imgConfig.AppName, imgConfig.Env)).Run()
	if err != nil {
		return fmt.Errorf("error getting state of current deployment: %s", err.Error())
	}

	err = exec.Command("terraform", "apply").Run()
	if err != nil {
		return fmt.Errorf("error exucuting terraform: %s", err.Error())
	}

	return nil
}
