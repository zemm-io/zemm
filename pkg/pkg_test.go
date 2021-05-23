package pkg

import (
	"fmt"
	"testing"
)

func TestZemmPkgNoUrl(t *testing.T) {
	_, err := NewPkg("https://hub.zemm.org/packages/library/nats/2.1.9/unknown.yaml")
	if err == nil {
		t.Error(fmt.Errorf("Loading packages from URL shouldn't be possible"))
	}

}

func TestParseZemmPkg(t *testing.T) {
	_, err := NewPkg("../examples/apps/minadmin/minadmin_pgsql/zemmpkg.yaml")
	if err != nil {
		t.Error(err)
	}
}

func TestVerifyZemmPkg(t *testing.T) {
	p, err := NewPkg("../examples/apps/minadmin/minadmin_pgsql/zemmpkg.yaml")
	if err != nil {
		t.Error(err)
	}

	err = p.Verify()
	if err != nil {
		t.Error(err)
	}
}

func TestMakePackage(t *testing.T) {
	p, err := NewPkg("../examples/apps/minadmin/minadmin_pgsql/zemmpkg.yaml")
	// p, err := NewPkg("../examples/apps/library/nats/zemmpkg.yaml")
	if err != nil {
		t.Error(err)
	}

	err = p.Verify()
	if err != nil {
		t.Error(err)
	}

	_, err = p.MakePackage()
	if err != nil {
		t.Error(err)
	}
}
