package schema

import (
	"encoding/json"
	"github.com/integration-system/isp-lib/utils"
	"os"
	p "path"
)

func ExtractConfig(path string) (map[string]interface{}, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]interface{})
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func ResolveDefaultConfigPath(path string) string {
	if utils.DEV {
		return p.Join("conf", path)
	}
	return path
}
