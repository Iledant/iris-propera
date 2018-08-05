package config

import (
	"testing"
)

func Test_getConfig(t *testing.T) {
	cfg := Get()
	if cfg == nil {
		t.Log("Impossible d'avoir la configuration")
		t.FailNow()
	}
	if cfg.Databases.Development.Name != "propera3" {
		t.Logf("Erreur sur le contenu de la configuration %+v", *cfg)
		t.Fail()
	}
}
