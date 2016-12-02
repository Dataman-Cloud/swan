package engine

import (
	"fmt"
)

func SubscribedHandler(h *Handler) *Handler {
	sub := h.MesosEvent.Event.GetSubscribed()
	fmt.Println(sub)

	return h
}
