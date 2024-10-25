package torrent

import (
	"bytes"
	"github.com/anacrolix/torrent/metainfo"
)

func StrDataToMagnetLink(data string) (string, error) {
	meta, err := metainfo.Load(bytes.NewReader([]byte(data)))
	if err != nil {
		return "", err
	}

	magnetLink, err := meta.MagnetV2()
	if err != nil {
		return "", err
	}

	return magnetLink.String(), nil
}
