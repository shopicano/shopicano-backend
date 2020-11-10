package queue

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/machinery"
	tasks2 "git.cloudbro.net/michaelfigg/yallawebsites/tasks"
	"github.com/RichardKnop/machinery/v1/tasks"
	"time"
)

func SendPaymentConfirmationEmail(orderID string) error {
	now := time.Now().Add(time.Second * 10)

	sig := &tasks.Signature{
		Name: tasks2.SendPaymentConfirmationEmailTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: orderID,
				Name:  "orderID",
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

func SendPaymentRevertedEmail(orderID string) error {
	now := time.Now().Add(time.Second * 10)

	sig := &tasks.Signature{
		Name: tasks2.SendPaymentRevertedEmailTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: orderID,
				Name:  "orderID",
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
