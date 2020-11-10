package queue

import (
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/shopicano/shopicano-backend/machinery"
	tasks2 "github.com/shopicano/shopicano-backend/tasks"
	"time"
)

func SendOrderDetailsEmail(orderID, subject string) error {
	now := time.Now().Add(time.Second * 10)

	sig := &tasks.Signature{
		Name: tasks2.SendOrderDetailsEmailTaskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: orderID,
				Name:  "orderID",
			},
			{
				Type:  "string",
				Value: subject,
				Name:  "subject",
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
