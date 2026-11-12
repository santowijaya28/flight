package config

type Config struct {
	MockDataPath MockDataPath
}

type MockDataPath struct {
	AirAsia         string
	BatikAir        string
	GarudaIndonesia string
	LionAir         string
}
