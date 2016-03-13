package main

import (
	"net/http"
	"strconv"
	"text/template"
)

// FormHandler - creates and returns a Handler for both Query and Form requests
func FormHandler(baseURL, confPath, confExt, mainDomain string,
	confTempl, resTempl template.Template,
	randHost <-chan string, done <-chan struct{}) http.HandlerFunc {

	confWriter := confWrite(confPath, confExt, mainDomain, confTempl, randHost)

	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.error(w, err.Error(), http.StatusBadRequest)
			r.Form = r.URL.Query()
		}
		var tls bool

		if len(r.Form["host"]) < 1 {
			http.Error(w, "no provided hostname", http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if len(r.Form["ip"]) < 1 {
			http.Error(w, "no provided ip", http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if tls, err = strconv.ParseBool(r.Form["tls"][0]); err != nil {
			tls = DefaultBackendTLS
		}

		config, err := confCheck(r.Form["host"][0], r.Form["ip"][0], tls,
			r.Form["blockedHeaders"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if config.ExtHost, err = confWriter(config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// TODO some log line? or no?
			return
		}

		if err = resTempl.Execute(w, config); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// TODO some long line? or no?
			return
		}
		return
	}
}

// JSONHandler - creates and returns a Handler for JSON body requests
func JSONHandler(baseURL, confPath, confExt, mainDomain string,
	confTempl, resTempl template.Template,
	randHost <-chan string, done <-chan struct{}) http.HandlerFunc {

	confWriter := confWrite(confPath, confExt, mainDomain, confTempl, randHost)

	return func(w http.ResponseWriter, r *http.Request) {

		var v []struct {
			host           string
			ip             string
			tls            bool
			blockedHeaders []string
		}

		err := json.Unmarshal(r.Body, &v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		for each := range v {
			config, err := confCheck(each.host, each.ip, each.tls, each.blockedHeaders)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				// TODO some log line?
				return
			}

			if config.ExtHost, err = confWriter(config); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				// TODO some log line? or no?
				return
			}

			if err = resTempl.Execute(w, config); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				// TODO some long line? or no?
				return
			}
			return

		}
	}
}
