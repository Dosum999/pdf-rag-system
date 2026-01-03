package config

import "os"

type Config struct {
	Database  DatabaseConfig
	DocReader DocReaderConfig
	Server    ServerConfig
	LLM       LLMConfig
	Embedding EmbeddingConfig
	Upload    UploadConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type DocReaderConfig struct {
	Host string
	Port string
}

type ServerConfig struct {
	Port string
	Host string
}

type LLMConfig struct {
	APIBaseURL string
	APIKey     string
	Model      string
}

type EmbeddingConfig struct {
	APIBaseURL string
	APIKey     string
	Model      string
	Dimension  int
}

type UploadConfig struct {
	Dir         string
	MaxFileSize int64
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "pdf_rag_db"),
		},
		DocReader: DocReaderConfig{
			Host: getEnv("DOCREADER_HOST", "localhost"),
			Port: getEnv("DOCREADER_PORT", "50051"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		LLM: LLMConfig{
			APIBaseURL: getEnv("LLM_API_BASE_URL", "https://api.openai.com/v1"),
			APIKey:     getEnv("LLM_API_KEY", ""),
			Model:      getEnv("LLM_MODEL", "gpt-4"),
		},
		Embedding: EmbeddingConfig{
			APIBaseURL: getEnv("EMBEDDING_API_URL", "https://api.openai.com/v1"),
			APIKey:     getEnv("EMBEDDING_API_KEY", ""),
			Model:      getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
			Dimension:  1536,
		},
		Upload: UploadConfig{
			Dir:         getEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSize: 50 * 1024 * 1024, // 50MB
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
