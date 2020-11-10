package queue

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/shopicano/shopicano-backend/machinery"
	tasks2 "github.com/shopicano/shopicano-backend/tasks"
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
