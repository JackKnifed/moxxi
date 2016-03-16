package moxxiConf

import (
	"github.com/dchest/uniuri"
	"net"
	"os"
	"strings"
	"text/template"
)

func inArr(a []string, t string) bool {
	for _, s := range a {
		if t == s {
			return true
		}
	}
	return false
}

// persistently runs and feeds back random URLs.
// To be started concurrently.
func RandSeqFeeder(baseURL string, excludes []string, length int,
	done <-chan struct{}) <-chan string {

	var feeder chan string
	if length < 2 {
		close(feeder)
		return feeder
	}

	go func() {
		var chars = []byte("abcdeefghijklmnopqrstuvwxyz")
		defer close(feeder)
		//rand.Seed(time.New().UnixNano())

		var newURL string

		for {
			newURL = uniuri.NewLenChars(length, chars) + "." + baseURL
			if inArr(excludes, newURL) {
				continue
			}
			select {
			case <-done:
				return
			case feeder <- newURL:
			}
		}
	}()

	return feeder
}

func validHost(s string) string {
	s = strings.Trim(s, ".")
	parts := strings.Split(s, DomainSep)
	if len(parts) < 2 {
		return ""
	}
	for i := 0; i < len(parts)-1; {
		if len(parts[i]) < 1 {
			parts = append(parts[:i], parts[i+1:]...)
		} else {
			i++
		}
	}
	return strings.Join(parts, DomainSep)
}

func confCheck(host, ip string, destTLS bool, port int, blockedHeaders []string) (siteParams, error) {
	var conf siteParams
	if conf.IntHost = validHost(host); conf.IntHost == "" {
		return siteParams{}, &Err{Code: ErrBadHost, value: host}
	}

	tempIP := net.ParseIP(ip)
	if tempIP == nil {
		return siteParams{}, &Err{Code: ErrBadIP, value: ip}
	}

	conf.IntPort = 80
	if port > 0 && port < MaxAllowedPort {
		conf.IntPort = port
	}

	conf.IntIP = tempIP.String()
	conf.Encrypted = destTLS
	conf.StripHeaders = blockedHeaders

	return conf, nil
}

func confWrite(confPath, confExt string, t template.Template,
	randHost <-chan string) func(siteParams) (siteParams, error) {

	return func(config siteParams) (siteParams, error) {

		err := os.ErrExist
		var randPart, fileName string
		var f *os.File

		for randPart == "" || os.IsExist(err) {
			select {
			case randPart = <-randHost:
			default:
				return siteParams{}, &Err{Code: ErrNoRandom}
			}
			fileName = strings.TrimRight(confPath, PathSep) + PathSep
			fileName += randPart + DomainSep + strings.TrimLeft(confExt, DomainSep)
			f, err = os.Create(fileName)
		}

		config.ExtHost = randPart

		if err == os.ErrPermission {
			return siteParams{}, &Err{Code: ErrFilePerm, value: fileName, deepErr: err}
		} else if err != nil {
			return siteParams{}, &Err{Code: ErrFileUnexpect, value: fileName, deepErr: err}
		}

		tErr := t.Execute(f, config)

		if err = f.Close(); err != nil {
			return siteParams{}, &Err{Code: ErrCloseFile, value: fileName, deepErr: err}
		}

		if tErr != nil {
			if err = os.Remove(fileName); err != nil {
				return siteParams{}, &Err{Code: ErrRemoveFile, value: fileName, deepErr: err}
			}
		}

		return config, nil
	}
}
