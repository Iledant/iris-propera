package config

import (
	"testing"
)

func Test_getConfig(t *testing.T) {
	var config ProperaConf
	if err := config.Get(); err != nil {
		t.Logf("Configuration, " + err.Error())
		t.FailNow()
	}
	if config.Databases.Development.Name != "propera3" {
		t.Logf("Erreur sur le contenu de la configuration %+v", config)
		t.Fail()
	}
}
