package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mdelio/redis-demo/backend"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

const (
	kStyle = `
{{define "STYLE"}}
<style>
table {
    font-family: arial, sans-serif;
    border-collapse: collapse;
    width: 100%;
}

td, th {
    border: 1px solid #dddddd;
    text-align: left;
    padding: 8px;
}

tr:nth-child(even) {
    background-color: #dddddd;
}
</style>
{{end}}
`

	kUserListTempl = `
<html><head>{{template "STYLE" .}}</head>
<body>
	{{if .UseEmoji}}
		<h1>&#x1f607;&#x1f607;</h1>
	{{end}}

	<table>
	<tr><th>Username</th></tr>
	{{range .UserList}}
		<tr><td><font color="blue">{{.}}</font></td></tr>
	{{else}}
		<tr><td><font color="red">None!</font></td></tr>
	{{end}}
	</table>
	{{fmtTime .Now}}
</body></html>`

	kUserInfoTempl = `
<html><head>{{template "STYLE" .}}</head>
<body>
	{{if .UseEmoji}}
		<h1>&#x1f608;&#x1f608;</h1>
	{{end}}

	<table>
	<tr><th>Username</th><th>Name</th><th>Password</th></tr>
	{{range $name := .SortedNames}}
		<tr><td>{{$name}}</td>
			{{$info := index $.Info $name}}
			<td><font color="blue">{{$info.Name}}</font></td>
			<td><font color="red">{{$info.Password}}</font></td>
		</tr>
	{{else}}
		<font color="red">None!</font>
	{{end}}
	</table>
	{{fmtTime .Now}}
</body></html>`
)

var (
	fRedisAddr    = flag.String("redis_addr", "localhost:6379", "Redis Address")
	fRedisTimeout = flag.Duration("redis_timeout", time.Millisecond*500, "Redis Dial Timeout")
	fListenAddr   = flag.String("listen_addr", "localhost:8080", "Listen Address")
	fProduction   = flag.Bool("is_production", false, "Use Production")
	fUseEmoji     = flag.Bool("print_emoji", false, "print emoji in the html")
)

type productionServer struct {
	tmpl     *template.Template
	client   *backend.Client
	useEmoji bool
}

func fmtTime(now time.Time) string {
	return now.Format("2006-01-02 15:04:05")
}

func newProductionServer(client *backend.Client, useEmoji bool) *productionServer {
	tmpl, err := template.New("list").
		Funcs(template.FuncMap{"fmtTime": fmtTime}).
		Parse(kUserListTempl + kStyle)
	if err != nil {
		log.Fatal(err)
	}

	return &productionServer{
		tmpl:     tmpl,
		client:   client,
		useEmoji: useEmoji,
	}
}

func (p *productionServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userNames, err := p.client.GetUserNames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	sort.Strings(userNames)
	data := struct {
		UseEmoji bool
		UserList []string
		Now      time.Time
	}{
		UseEmoji: p.useEmoji,
		UserList: userNames,
		Now:      time.Now(),
	}

	if err := p.tmpl.Execute(w, data); err != nil {
		err = fmt.Errorf("failed to execute html template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

type developmentServer struct {
	tmpl     *template.Template
	client   *backend.Client
	useEmoji bool
}

func newDevelopmentServer(client *backend.Client, useEmoji bool) *developmentServer {
	tmpl, err := template.New("list").
		Funcs(template.FuncMap{"fmtTime": fmtTime}).
		Parse(kUserInfoTempl + kStyle)
	if err != nil {
		log.Fatal(err)
	}

	return &developmentServer{
		tmpl:     tmpl,
		client:   client,
		useEmoji: useEmoji,
	}
}

func (d *developmentServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userInfoMap, err := d.client.GetAllUserInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	sortedNames := make([]string, len(userInfoMap))
	var idx int
	for name := range userInfoMap {
		sortedNames[idx] = name
		idx++
	}
	sort.Strings(sortedNames)

	data := struct {
		UseEmoji    bool
		SortedNames []string
		Info        map[string]backend.UserInfo
		Now         time.Time
	}{
		UseEmoji:    d.useEmoji,
		SortedNames: sortedNames,
		Info:        userInfoMap,
		Now:         time.Now(),
	}

	if err := d.tmpl.Execute(w, data); err != nil {
		err = fmt.Errorf("failed to execute html template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func main() {
	log.Printf("flags: %v", os.Args)
	flag.Parse()

	conf := struct {
		Production   bool
		RedisAddr    string
		RedisTimeout time.Duration
		ListenAddr   string
		UseEmoji     bool
	}{
		Production:   *fProduction,
		RedisAddr:    *fRedisAddr,
		RedisTimeout: *fRedisTimeout,
		ListenAddr:   *fListenAddr,
		UseEmoji:     *fUseEmoji,
	}
	confJson, err := json.Marshal(conf)
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}
	log.Printf("config: %v\n", confJson)

	client := backend.NewClient(conf.RedisAddr, conf.RedisTimeout)

	var handler http.Handler
	if conf.Production {
		err := client.SeedData(map[string]backend.UserInfo{
			"anaim":  {Name: "Allan", Password: "my-super-secret-password"},
			"mdelio": {Name: "Matthew", Password: "my-password"},
		})
		if err != nil {
			log.Fatalf("failed to insert info: %v", err)
		}
		log.Println("setting up production server")
		handler = newProductionServer(client, conf.UseEmoji)
	} else {
		log.Println("setting up development server")
		handler = newDevelopmentServer(client, conf.UseEmoji)
	}

	log.Printf("Listening on %q\n", conf.ListenAddr)
	if err := http.ListenAndServe(conf.ListenAddr, handler); err != nil {
		log.Fatal(err)
	}
}
