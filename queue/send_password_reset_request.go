package queue

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/shopicano/shopicano-backend/machinery"
	tasks2 "github.com/shopicano/shopicano-backend/tasks"
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
