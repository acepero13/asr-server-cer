package config

import (
	config2 "cloud-client-go/config"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
)

type configPool struct {
	configurations map[*config2.Config]bool
	poolMutex      *sync.Mutex
	maxConfigs     int
}

var configurationPool = &configPool{
	maxConfigs:     10,
	poolMutex:      &sync.Mutex{},
	configurations: make(map[*config2.Config]bool),
}

func init() {
	configBaseName := "asr_sem_$.json"
	baseConfigPath := getBaseConfigPath()
	for i := 0; i < configurationPool.maxConfigs; i++ {
		configPath := baseConfigPath + strings.Replace(configBaseName, "$", strconv.Itoa(i), 1)
		jsonConfig := config2.ReadConfig(configPath)
		configurationPool.configurations[jsonConfig] = false

	}
}

//GiveMeAConfig Returns a config from the config pool
func GiveMeAConfig() (*config2.Config, error) {
	configurationPool.poolMutex.Lock()
	defer configurationPool.poolMutex.Unlock()
	for config, isUsed := range configurationPool.configurations {
		if !isUsed {
			configurationPool.configurations[config] = true
			return config, nil
		}
	}
	return nil, errors.New("config pool is full. I cannot provide you with one")
}

//Release Releases an already used config file so it can be re-used again
func Release(config *config2.Config) error {
	configurationPool.poolMutex.Lock()
	defer configurationPool.poolMutex.Unlock()
	if _, ok := configurationPool.configurations[config]; ok {
		configurationPool.configurations[config] = false
		return nil
	}
	return errors.New("cannot find the given config")
}

func getBaseConfigPath() string {
	isTestingEnvironment := strings.HasSuffix(os.Args[0], ".test")
	if isTestingEnvironment {
		return "../../configs/"
	}
	return "configs/"

}
