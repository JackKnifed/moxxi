package main

import (
	"net/http"
	"text/template"
	"strconv"
)

func QueryHandler(baseURL, confPath, confExt, mainDomain string,
	confTempl, resTempl template.Template,
	randHost <-chan string, done <-chan struct{}) http.HandlerFunc {

	confWriter := confWrite(confPath, confExt, mainDomain, confTempl, randHost)

	return func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		var tls bool
		var err error

		if len(values["host"]) < 1 {
			http.Error(w, "no provided hostname", http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if len(values["ip"]) < 1 {
			http.Error(w, "no provided ip", http.StatusPreconditionFailed)
			// TODO some log line?
			return
		}

		if tls, err = strconv.ParseBool(values["tls"][0]); err != nil {
			tls = DefaultBackendTLS
		}

		config, err := confCheck(values["host"][0], values["ip"][0], tls,
			values["blockedHeaders"])
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

func FormHandler(baseURL, confPath, confExt string, subdomainLength int,
	confTempl, resTempl template.Template, randHost <-chan string,
	done <-chan struct{}) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// r.ParseForm
	}
}
