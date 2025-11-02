package hibikecustomsource

type (
	Settings struct {
		SupportsAnime bool `json:"supportsAnime"`
		SupportsManga bool `json:"supportsManga"`
	}

	Provider interface {
	}
)
