package config

import (
	config2 "cloud-client-go/config"
	"github.com/acepero13/asr_server/server/config"
	"testing"
)

func Test_configPool_GiveMeAConfig(t *testing.T) {
	c, err := config.GiveMeAConfig()

	if err != nil {
		t.Errorf("Error occurred")
	}
	if c == nil || c.Port != 443 {
		t.Errorf("Port not recognized")
	}

	erro := config.Release(c)
	if erro != nil {
		t.Errorf("Could not release config")
	}

}

func Test_configPool_GiveMeAConfigTwoTimesAndItShouldBeDifferent(t *testing.T) {

	c1, _ := config.GiveMeAConfig()
	boundary1 := c1.GetBoundary()
	c, err := config.GiveMeAConfig()

	if c == nil || c.GetBoundary() == boundary1 {
		t.Errorf("id is wrong")
	}

	if err != nil {
		t.Errorf("Error occurred")
	}
	if c == nil || c.Port != 443 {
		t.Errorf("Port not recognized")
	}

	_ = config.Release(c1)
	_ = config.Release(c)

}

func Test_configPool_GiveMeMaxConfigsTheNextOneGivesError(t *testing.T) {

	var configs []*config2.Config
	configs = make([]*config2.Config, 0, 10)
	for i := 0; i < 10; i++ {
		c, _ := config.GiveMeAConfig()
		configs = append(configs, c)
	}

	_, err := config.GiveMeAConfig()
	if err == nil {
		t.Errorf("The pool should be full")
	}

	for _, c := range configs {
		_ = config.Release(c)
	}

}

func Test_configPool_GiveMeMaxConfigsAfterRelseasingICanGetOneBack(t *testing.T) {

	var last *config2.Config
	for i := 0; i < 10; i++ {
		last, _ = config.GiveMeAConfig()
	}

	erro := config.Release(last)

	if erro != nil {
		t.Errorf("Should not raise error. " + erro.Error())
	}
	_, err := config.GiveMeAConfig()
	if err != nil {
		t.Errorf("One should be free")
	}

}
