package queue

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/machinery"
	tasks2 "git.cloudbro.net/michaelfigg/yallawebsites/tasks"
	"github.com/RichardKnop/machinery/v1/tasks"
	"time"
)

func SendPasswordResetRequestEmail(userID string) error {
	now := time.Now().Add(time.Second * 10)

	sig := &tasks.Signature{
		Name: tasks2.SendResetPasswordEmailTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: userID,
				Name:  "userID",
			},
		},
		ETA: &now,
	}
	_, err := machinery.RabbitMQConnection().SendTask(sig)
	if err != nil {
		return err
	}
	return nil
}

func SendPasswordResetConfirmationEmail(userID string) error {
	now := time.Now().Add(time.Second * 10)

	sig := &tasks.Signature{
		Name: tasks2.SendResetPasswordConfirmationEmailTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: userID,
				Name:  "userID",
			},
		},
		ETA: &now,
	}
	_, err := machinery.RabbitMQConnection().SendTask(sig)
	if err != nil {
		return err
	}
	return nil
}
