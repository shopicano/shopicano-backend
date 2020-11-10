package queue

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/machinery"
	tasks2 "git.cloudbro.net/michaelfigg/yallawebsites/tasks"
	"github.com/RichardKnop/machinery/v1/tasks"
)

func SendSignUpVerificationEmail(userID string) error {
	sig := &tasks.Signature{
		Name: tasks2.SendSignUpVerificationEmailTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: userID,
				Name:  "userID",
			},
		},
	}
	_, err := machinery.RabbitMQConnection().SendTask(sig)
	if err != nil {
		return err
	}
	return nil
}
