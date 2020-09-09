package queue

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/machinery"
	tasks2 "git.cloudbro.net/michaelfigg/yallawebsites/tasks"
	"github.com/RichardKnop/machinery/v1/tasks"
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
