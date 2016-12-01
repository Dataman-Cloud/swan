package middleware

import (
	"github.com/Dataman-Cloud/swan/src/manager/framework/event/handler"
)

func SubscribedHandler(h *handler.Handler) *handler.Handler {
	sub := event.GetSubscribed()

	return h
}
