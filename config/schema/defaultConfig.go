package schema

import (
	"encoding/json"
	"github.com/integration-system/isp-lib/logger"
	"os"
)

func ExtractConfig(path string) map[string]interface{} {
	if path == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Error(err)
		return nil
	}
	defer file.Close()

	config := make(map[string]interface{})
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		logger.Error(err)
		return nil
	}
	return config
}
