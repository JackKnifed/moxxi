package main

import (
	"flag"
	"github.com/JackKnifed/moxxi/moxxiconf"
	"log"
	"net/http"
	"io/ioutil"
	"text/template"
)

func main() {
	var jsonHandler moxxiConf.HandlerLocFlag
	var formHandler moxxiConf.HandlerLocFlag
	var fileHandler moxxiConf.HandlerLocFlag
	var fileDocroot moxxiConf.HandlerLocFlag
	var excludedDomains moxxiConf.HandlerLocFlag
	var staticHandler moxxiConf.HandlerLocFlag
	var staticResponse moxxiConf.HandlerLocFlag

	listen := flag.String("listen", ":8080", "listen address to use")
	confTemplString := flag.String("confTempl", "template.conf", "base templates for the configs")
	resTemplString := flag.String("resTempl", "template.response", "base template for the response")
	baseDomain := flag.String("domain", "", "base domain to add onto")
	subdomainLength := flag.Int("subLength", 8, "length of subdomain to exclude")
	confLoc := flag.String("confLoc", ".", "path to put the domains")
	confExt := flag.String("confExt", ".conf", "extension to add to the confs")

	flag.Var(&excludedDomains, "excludedDomain", "domain names to exclude")
	flag.Var(&jsonHandler, "jsonHandler", "locations for a JSON handler (multiple)")
	flag.Var(&formHandler, "formHandler", "locations for a form handler (multiple)")
	flag.Var(&fileHandler, "fileHandler", "locations for a file handler (multiple)")
	flag.Var(&fileDocroot, "fileDocroot", "docroots for each file handler (multiple)")
	flag.Var(&staticHandler, "staticHandler", "location for a static response (multiple)")
	flag.Var(&staticResponse, "staticResponse", "file containing the static response to return (multiple)")

	flag.Parse()

	confTempl, err := template.ParseFiles(*confTemplString)
	if err != nil {
		log.Fatal(err)
	}
	resTempl, err := template.ParseFiles(*resTemplString)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	for i, _ := range staticResponse {
		newResponse, err := ioutil.ReadFile(staticResponse[i])
		if err != nil {
			log.Fatal(err)
		}
		mux.HandleFunc(staticHandler[i], moxxiConf.StaticHandler(newResponse))
	}

	for _, each := range jsonHandler {
		mux.HandleFunc(each, moxxiConf.JSONHandler(*baseDomain, *confLoc, *confExt, excludedDomains, *confTempl, *resTempl, *subdomainLength))
	}

	for _, each := range formHandler {
		mux.HandleFunc(each, moxxiConf.FormHandler(*baseDomain, *confLoc, *confExt, excludedDomains, *confTempl, *resTempl, *subdomainLength))
	}

	if len(fileHandler) != len(fileDocroot) {
		log.Fatal("mismatch between docroots and filehandlers")
	}

	for i := 0; i < len(fileHandler); i++ {
		mux.Handle(fileHandler.GetOne(i), http.FileServer(http.Dir(fileDocroot.GetOne(i))))
	}

	srv := http.Server{
		Addr:         *listen,
		Handler:      mux,
		ReadTimeout:  moxxiConf.ConnTimeout,
		WriteTimeout: moxxiConf.ConnTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}
