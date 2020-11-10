package queue

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/shopicano/shopicano-backend/machinery"
	tasks2 "github.com/shopicano/shopicano-backend/tasks"
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
