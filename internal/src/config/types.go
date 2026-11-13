package config

type Config struct {
	MockDataPath MockDataPath
	Cache        CacheConfig
	Provider     ProviderConfig
}

type MockDataPath struct {
	AirAsia         string
	BatikAir        string
	GarudaIndonesia string
	LionAir         string
}

type CacheConfig struct {
	MaxEntries int
	TTLSeconds int
}

type ProviderConfig struct {
	TimeoutMs        int
	MaxAttempts      int
	InitialBackoffMs int
	MaxBackoffMs     int
}
