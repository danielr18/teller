package providers

import (
	"testing"

	"github.com/alecthomas/assert"

	"github.com/danielr18/teller/pkg/core"
	"github.com/danielr18/teller/pkg/logging"
)

func AssertProvider(t *testing.T, s core.Provider, sync bool) {
	p := core.NewPopulate(map[string]string{"stage": "prod"})
	kpmap := p.KeyPath(core.KeyPath{Field: "MG_KEY", Path: "settings/{{stage}}/billing-svc/all", Decrypt: true})
	kp := p.KeyPath(core.KeyPath{Field: "MG_KEY", Path: "settings/{{stage}}/billing-svc", Decrypt: true})
	kpenv := p.KeyPath(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc", Decrypt: true})

	ent, err := s.Get(kp)
	assert.Nil(t, err)
	assert.Equal(t, ent.Value, "shazam")

	ent, err = s.Get(kpenv)
	assert.Nil(t, err)
	assert.Equal(t, ent.Value, "shazam")

	if sync {
		ents, err := s.GetMapping(kpmap)
		assert.Nil(t, err)
		assert.Equal(t, len(ents), 2)
		assert.Equal(t, ents[0].Value, "mailman")
		assert.Equal(t, ents[1].Value, "shazam")
	}
}

func GetTestLogger() logging.Logger {
	logger := logging.New()
	logger.SetLevel("null")
	return logger
}
