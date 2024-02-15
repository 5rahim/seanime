package updater

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdater_getReleaseName(t *testing.T) {

	updater := Updater{}

	t.Log(updater.getReleaseName("0.2.2"))
}

func TestUpdater_FetchLatestRelease(t *testing.T) {

	updater := Updater{}
	release, err := updater.fetchLatestRelease()
	if err != nil {
		t.Fatal(err)
	}

	if assert.NotNil(t, release) {
		spew.Dump(release)
	}

}
