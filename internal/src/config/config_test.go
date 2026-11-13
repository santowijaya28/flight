package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func stringPtr(value string) *string {
	return &value
}

func validConfigContent() string {
	return `[MockDataPath]
AirAsia = "external/airasia_search_response.json"
BatikAir = "external/batik_air_search_response.json"
GarudaIndonesia = "external/garuda_indonesia_search_response.json"
LionAir = "external/lion_air_search_response.json"

[Cache]
MaxEntries = 10000
TTLSeconds = 30

[Provider]
TimeoutMs = 1000
MaxAttempts = 3
InitialBackoffMs = 50
MaxBackoffMs = 250
`
}

func TestInitConfig(t *testing.T) {
	want := Config{
		MockDataPath: MockDataPath{
			AirAsia:         "external/airasia_search_response.json",
			BatikAir:        "external/batik_air_search_response.json",
			GarudaIndonesia: "external/garuda_indonesia_search_response.json",
			LionAir:         "external/lion_air_search_response.json",
		},
		Cache: CacheConfig{
			MaxEntries: 10000,
			TTLSeconds: 30,
		},
		Provider: ProviderConfig{
			TimeoutMs:        1000,
			MaxAttempts:      3,
			InitialBackoffMs: 50,
			MaxBackoffMs:     250,
		},
	}

	tests := []struct {
		name    string
		content *string
		wantErr bool
	}{
		{
			name:    "valid config",
			content: stringPtr(validConfigContent()),
		},
		{
			name:    "missing config file",
			content: nil,
			wantErr: true,
		},
		{
			name:    "invalid ini syntax",
			content: stringPtr("[MockDataPath\nAirAsia = \"file.json\""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldWD, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			oldCfg := cfg
			t.Cleanup(func() {
				cfg = oldCfg
				_ = os.Chdir(oldWD)
			})

			root := t.TempDir()
			configDir := filepath.Join(root, "internal", "files", "etc", "conf")
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				t.Fatal(err)
			}

			if tt.content != nil {
				path := filepath.Join(configDir, "config.ini")
				if err := os.WriteFile(path, []byte(*tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			if err := os.Chdir(root); err != nil {
				t.Fatal(err)
			}
			cfg = Config{}

			got, err := InitConfig()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("config is nil")
			}
			if !reflect.DeepEqual(*got, want) {
				t.Fatalf("config = %+v, want %+v", *got, want)
			}
			if !reflect.DeepEqual(*GetConfig(), want) {
				t.Fatalf("GetConfig() = %+v, want %+v", *GetConfig(), want)
			}
		})
	}
}
