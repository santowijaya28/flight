package provider

type provider struct {
	providers map[string]Provider
}

func Init(providers map[string]Provider) *provider {
	return &provider{
		providers: providers,
	}
}
