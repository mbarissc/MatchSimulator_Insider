// FILE: config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"log" // Loglama için
	"os"
)

// Config struct, config.json dosyasındaki tüm yapılandırmayı tutar.
type Config struct {
	Database DBConfig  `json:"database"` // mapstructure yerine json etiketi
	Server   APIConfig `json:"server"`   // mapstructure yerine json etiketi
}

// DBConfig, veritabanı bağlantı ayarlarını tutar.
type DBConfig struct {
	ConnectionString string `json:"connectionString"`
}

// APIConfig, API sunucusu ayarlarını tutar.
type APIConfig struct {
	Port string `json:"port"`
}

// AppConfig, yüklenen yapılandırmayı tutacak global değişken (isteğe bağlı).
// Eğer global kullanmak istemiyorsanız, LoadConfig'den dönen değeri main içinde kullanırsınız.
var AppConfig Config

// LoadConfig, belirtilen yoldan yapılandırmayı (config.json) yükler.
func LoadConfig(filePath string) (*Config, error) {
	log.Printf("INFO: Loading configuration from %s", filePath)

	configFile, err := os.Open(filePath)
	if err != nil {
		// Dosya bulunamazsa veya açılamazsa, varsayılan değerleri kullanmayı veya hata vermeyi seçebilirsiniz.
		// Bu örnekte, varsayılan değerleri kullanarak devam edip bir uyarı loglayacağız.
		log.Printf("WARNING: Could not open config file '%s': %v. Using default values.", filePath, err)
		// Varsayılan değerleri burada set edebilirsiniz veya main.go'da kontrol edebilirsiniz.
		// Şimdilik, main.go'daki gibi varsayılanlara düşmesini sağlayalım.
		// Ya da burada direkt varsayılan bir Config objesi döndürebiliriz.
		// Basitlik adına, hata döndürelim ve main.go'da bu hatayı yakalayıp varsayılanları kullanalım.
		return nil, fmt.Errorf("could not open config file '%s': %w", filePath, err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	var cfg Config
	if err = decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decoding config file '%s': %w", filePath, err)
	}

	// Port için varsayılan değer kontrolü (eğer dosyada yoksa veya boşsa)
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080" // Varsayılan port
		log.Println("INFO: API port not found in config, using default '8080'.")
	}
	// ConnectionString için varsayılan değer kontrolü
	if cfg.Database.ConnectionString == "" {
		// Bu durumda programın çalışması zor olacağı için hata vermek daha mantıklı olabilir
		// veya main.go'da bir fallback sağlanabilir. Şimdilik dosyada olmasını bekleyelim.
		log.Println("WARNING: Database connectionString not found in config. Application might not connect to DB.")
		// return nil, fmt.Errorf("database connectionString is missing in config file")
	}

	AppConfig = cfg // Global değişkene ata (isteğe bağlı)
	return &cfg, nil
}
