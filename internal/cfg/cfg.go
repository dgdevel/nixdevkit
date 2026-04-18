package cfg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const FileName = ".nixdevkitrc"

func FilePath(rootDir string) string {
	return rootDir + "/" + FileName
}

func Read(path string) (map[string]map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]map[string]string{}, nil
		}
		return nil, err
	}
	return Parse(string(data)), nil
}

func Parse(data string) map[string]map[string]string {
	config := map[string]map[string]string{}
	var section string
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			if config[section] == nil {
				config[section] = map[string]string{}
			}
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			if config[section] == nil {
				config[section] = map[string]string{}
			}
			config[section][key] = val
		}
	}
	return config
}

func Write(config map[string]map[string]string, path string) error {
	var buf strings.Builder
	for section, keys := range config {
		buf.WriteString(fmt.Sprintf("[%s]\n", section))
		for key, val := range keys {
			buf.WriteString(fmt.Sprintf("%s=%s\n", key, val))
		}
	}
	return os.WriteFile(path, []byte(buf.String()), 0644)
}

func ParseBool(val string) bool {
	switch strings.ToLower(val) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func IsReadonly(rootDir string) bool {
	config, err := Read(FilePath(rootDir))
	if err != nil {
		return false
	}
	if core, ok := config["core"]; ok {
		return ParseBool(core["readonly"])
	}
	return false
}
