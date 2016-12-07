package scheduler

import ()

func DummyHandler(h *Handler) (*Handler, error) {
	return h, nil
}
