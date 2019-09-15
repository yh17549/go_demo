package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var workDir string

func handerListFile(w http.ResponseWriter, r *http.Request) {

	dirList, e := ioutil.ReadDir(workDir)
	if e != nil {
		fmt.Println("read dir error")
		return
	}

	var fileListTemplate = template.
		Must(template.New("filelist").
			Parse(`
<h1>file list</h1>
<table>
	<tr style='text-align: left'>
		<th>Title</th>
	</tr>
	{{range .}}
	<tr>
		<td> <a href='file/{{.Name}}'>{{.Name}}</a> </td>
	</tr>
	{{end}}
</table>
`))
	fileListTemplate.Execute(w, dirList)
}

var port string
var autoShutdown bool
var shutdownMinute int

func init() {
	flag.StringVar(&port, "port", "8809", "port")
	flag.BoolVar(&autoShutdown, "autoShutdown", true, "application will shutdown after a specified time")
	flag.IntVar(&shutdownMinute, "shutdownMinute", 5, "the specified time (minute)")
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage :\n")
		flag.PrintDefaults()
		fmt.Println("Example:")
		fmt.Println("  sh ./shareFile -port 8890 -shutdownMinute 1 -autoShutdown=false")
	}
	flag.Parse()

	var err error
	workDir, err = os.Getwd()
	if err != nil {
		fmt.Println("get work dir error", err)
		return
	}

	fmt.Println("share dir:", workDir)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return
	}
	http.HandleFunc("/", handerListFile)
	http.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir(workDir))))
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				listenURL := ipnet.IP.String() + ":" + port
				fmt.Println("please visit: http://" + listenURL)
				go func(listenURL string) {

					log.Fatal(http.ListenAndServe(listenURL, nil))
				}(listenURL)
			}
		}
	}

	fmt.Println("file is sharing...")

	if autoShutdown {
		now := time.Now()

		next := now.Add(time.Minute * time.Duration(shutdownMinute))

		t := time.NewTimer(next.Sub(now))
		<-t.C
		fmt.Println("auto close application after ", shutdownMinute, " minutes")
	} else {
		select {}
	}

}
