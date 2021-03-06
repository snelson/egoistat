package main

import (
	"encoding/json"
	"flag"
	"github.com/nu7hatch/egoistat/backend"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

var Addr string

func init() {
	flag.StringVar(&Addr, "addr", ":8080", "The address to serve on")
	flag.Parse()
}

var scriptTpl = "var egoistat={c:{{.Data}}, points: function(sn){return this.c[sn] || 0;}};\n{{.Callback}};"

func transformParams(form url.Values) (res map[string]string) {
	res = make(map[string]string)
	for k, v := range form {
		res[k] = strings.Join(v, "\n")
	}
	return
}

func statHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var networks = strings.Split(r.FormValue("n"), ",")
	var url = r.FormValue("url")
	var params = transformParams(r.Form)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	request := egoistat.NewRequest(url, params)
	results := request.Stat(networks...)

	enc := json.NewEncoder(w)
	enc.Encode(results)
}

type countScriptData struct {
	Data     string
	Callback string
}

func statScriptHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var networks = strings.Split(r.FormValue("n"), ",")
	var url = r.FormValue("url")
	var params = transformParams(r.Form)
	var callback = r.FormValue("cb")

	if len(strings.TrimSpace(callback)) > 0 {
		callback = callback + "()"
	}

	w.Header().Set("Content-Type", "text/javascript")
	w.WriteHeader(200)

	request := egoistat.NewRequest(url, params)
	results := request.Stat(networks...)
	data, _ := json.Marshal(results)

	tmpl, _ := template.New("script").Parse(scriptTpl)
	tmpl.Execute(w, countScriptData{string(data), callback})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./public/index.html")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/stat/", indexHandler)
	http.HandleFunc("/api/v1/stat.json", statHandler)
	http.HandleFunc("/api/v1/stat.js", statScriptHandler)

	log.Fatal(http.ListenAndServe(Addr, nil))
}
