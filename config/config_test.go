package config

import (
	"testing"
)

func Test_getConfig(t *testing.T) {
	var config ProperaConf
	if err := config.Get(); err != nil {
		t.Logf("Récupération de la configuration : " + err.Error())
		t.FailNow()
	}
	if config.Databases.Development.Name != "propera3" {
		t.Logf("Contenu de la configuration : %+v", config)
		t.Fail()
	}
}
