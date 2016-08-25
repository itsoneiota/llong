package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/abourget/slick"
	_ "github.com/abourget/slick/web"
	// _ "github.com/abourget/slick/webauth"
	_ "github.com/itsoneiota/llong/buttons"
	_ "github.com/itsoneiota/llong/deploybot"
	"github.com/owenmorgan/llongdocker"
)

var configFile = flag.String("config", "slick.conf", "config file")

func main() {
	flag.Parse()

	bot := slick.New(*configFile)

	bot.Run()
}

//Deploy deploys the app
func Deploy(image, tag, env string) error {

	client := llongClient()

	imgConfig, err := client.GetImageConfig(image, tag)
	imgConfig.Image = image
	imgConfig.Env = env

	t := template.New("app")

	t, err = t.Parse(`
    provider "aws"{
        region = "eu-west-1"
    }
    
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

	t.Execute(f, imgConfig)

	err = exec.Command("terraform", "apply").Run()
	if err != nil {
		return fmt.Errorf("error exucuting terraform: %s", err.Error())
	}

	return nil
}

func llongClient() *llongdocker.LlongDockerClient {
	return llongdocker.NewLlongDockerClient(
		"eu-west-1",
		"unix:///var/run/docker.sock",
		"https://597304777786.dkr.ecr.eu-west-1.amazonaws.com",
		"AWS",
		"AQECAHh+dS+BlNu0NxnXwowbILs115yjd+LNAZhBLZsunOxk3AAAAvEwggLtBgkqhkiG9w0BBwagggLeMIIC2gIBADCCAtMGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMd90K7Qe01TJN2HZwAgEQgIICpIcbw6UQlHWwwxB/ibJV2t+rAuCfWLL41sRKdfjrpHTC+/YVkMA3679bCqsg2XPE73i6offn5DLW5/KjljySpFvq0jK/8n0su9xSki1NV4D6OBhpBCY4wMQXc+qOXKVj1/5BYJoo1TDE8hxVf946fkAQctXfUEN9vdnxreknPGaNC1UlIlLGZ2xhPiZ0t7cjiXzk1wpwsocf39SSHHF7LYbrEkMFhEXBsTqCP7zx2uo9AFonYMFA6fqmWELyansPVvXytRHF3g7iMY/8++uWLMWdsq9I/Waproosphzl4fUSmOobzomp09x1eOfZvdyI4WRKGFFsHG4T4mWB1vSbJIDVl0lr64LcwjauE0IlcXgtdPvv5lMAjz+tqdZ6MnqQkZMObXLx8xnzyHRjnTx3SwmBTfHatFVeK9RC3pW9DTIYfStVTgZOdbDgZeuadkWD8fJqW+hH3XtCkOnBnDaeyOTnWnnS+ANWfLaUy2tWFRmDnlcOOkB2riNA70hYkmj+aXRJF52pi6HEDMUqp+zhCSGfew82FSKc9nFDAX2pyod1M9gxuwJC6FReyiimyHH7jT+qiVSvntPMBABWZWuZBbH7/iIZKPIsiHth//7bS71pMwim5HjJLPB+WFHuC1/xEbwkQmzmbpJWFmWDyST6H6VuCHUxes/mZKoSycWTtDqPwTyyU54xvSiIX/zT3XI8AXxdtDp5afbkADhwy9950DO7P1K5XxBKGPYzB18TLu7dytHBDp8HCt5YaxAVTNqR7VJZjUGyt3uXCCunJMO/Ktkq8xN53u+L8X3JiBsh+G90DqS7A5DEEcpTlkztT/OYCERYVblLFZYOFpBARSaea4NXHLjUT5wHlniAZ4AO76u6M4hV4bNJYDydyKbM6isIazXuYIE=",
	)
}
