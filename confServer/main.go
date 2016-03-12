package main

import (
	"github.com/dchest/uniuri"
	"net"
	"os"
	"strings"
	"text/template"
)

// persistently runs and feeds back random URLs.
// To be started concurrently.
func randSeqFeeder(baseURL string, length int, feeder chan<- string, done <-chan struct{}) {

	var chars = []byte("abcdeefghijklmnopqrstuvwxyz")
	defer close(feeder)
	//rand.Seed(time.New().UnixNano())
	newURL := uniuri.NewLenChars(length, chars) + "." + baseURL

	for {
		select {
		case <-done:
			return
		case feeder <- newURL:
			newURL = uniuri.NewLenChars(length, chars) + "." + baseURL
		}
	}
}

func validHost(s string) string {
	s = strings.Trim(s, ".")
	parts := len(strings.Split(s, "."))
	if parts < 2 {
		return ""
	}
	return s
}

func configCheck(host, ip string, destTLS bool, blockedHeaders []string) (siteParams, error) {
	var conf siteParams
	if conf.IntHost = validHost(host); conf.IntHost == "" {
		return siteParams{}, &Err{Code: ErrBadHost, value: ip}
	}

	tempIP := net.ParseIP(ip)
	if tempIP == nil {
		return siteParams{}, &Err{Code: ErrBadIP, value: ip}
	}

	conf.IntIP = tempIP.String()
	conf.Encrypted = destTLS
	conf.StripHeaders = blockedHeaders

	return conf, nil
}

func writeConf(config siteParams, confPath, confExt string, t template.Template, randHost <-chan string) (string, error) {

	// set up the filename once
	config.ExtHost = <-randHost
	fileName := strings.TrimRight(confPath, PathSep)
	fileName += PathSep + config.ExtHost + confExt
	out, err := os.Create(fileName)

	// if you get an filename exists error, keep doing it until you don't
	for err == nil && os.IsExist(err) {
		config.ExtHost = <-randHost
		fileName = strings.TrimRight(confPath, PathSep)
		fileName += PathSep + config.ExtHost + confExt
		out, err = os.Create(fileName)
	}

	if err == os.ErrPermission {
		return "", &Err{Code: ErrFilePerm, value: fileName, deepErr: err}
	}

	if err != nil {
		return "", &Err{Code: ErrFileUnexpect, value: fileName, deepErr: err}
	}

	tErr := t.Execute(out, config)

	if err = out.Close(); err != nil {
		return "", &Err{Code: ErrCloseFile, value: fileName, deepErr: err}
	}

	if tErr != nil {
		if err = os.Remove(fileName); err != nil {
			return "", &Err{Code: ErrRemoveFile, value: fileName, deepErr: err}
		}
	}

	return config.ExtHost, nil
}
