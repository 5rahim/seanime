package animetosho

import "testing"

func TestMagnet(t *testing.T) {

	url := "https://animetosho.org/view/kaizoku-jujutsu-kaisen-26-a1c9bab1-season-2.n1710116"

	magnet, err := TorrentMagnet(url)

	if err != nil {
		t.Fatal(err)
	}

	if magnet == "" {
		t.Fatal("magnet link not found")
	}

	t.Log(magnet)

}

func TestTorrentFile(t *testing.T) {

	url := "https://animetosho.org/view/kaizoku-jujutsu-kaisen-26-a1c9bab1-season-2.n1710116"

	magnet, err := TorrentFile(url)

	if err != nil {
		t.Fatal(err)
	}

	if magnet == "" {
		t.Fatal("download link not found")
	}

	t.Log(magnet)

}
