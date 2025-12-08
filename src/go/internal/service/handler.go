package service

// handler implements the main logic for processing incoming events.
type handler struct {
	cache *Cache
	repo  Repository
}

func NewHandler(c *Cache, r Repository) Handler {
	return &handler{cache: c, repo: r}
}

// Handle processes a single balance-change event.
func (h *handler) Handle(e Event) error {
	updated, err := h.repo.ApplyEvent(e)
	if err != nil {
		return err
	}
	if !updated {
		// duplicate or stale event
		return nil
	}

	h.cache.SetIfNewer(e.UserID, e.WalletID, e.WalletBalance, e.UpdatedAt)
	return nil
}
