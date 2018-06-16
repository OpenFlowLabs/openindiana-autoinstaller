package fileutils

import (
	"time"

	"git.wegmueller.it/toasterson/glog"
	"github.com/cavaliercoder/grab"
)

func HTTPDownload(url string, location string) (err error) {
	if location == "" {
		location = "/tmp/"
	}
	req, _ := grab.NewRequest(location, url)
	return doDownload(req).Err()
}

func HTTPDownloadTo(url string, location string) (file string, err error) {
	if location == "" {
		location = "/tmp/"
	}
	req, _ := grab.NewRequest(location, url)
	resp := doDownload(req)
	return resp.Filename, resp.Err()
}

func doDownload(request *grab.Request) (resp *grab.Response) {
	client := grab.NewClient()
	// start download
	glog.Infof("Downloading %v...", request.URL())
	resp = client.Do(request)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

ProgressLoop:
	for {
		select {
		case <-t.C:
			glog.Infof("  transferred %v / %v bytes (%.2f%%)",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break ProgressLoop
		}
	}

	glog.Infof("Download saved to %v", resp.Filename)
	return
}
