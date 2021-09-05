package main

import (
        "flag"
        "fmt"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "log"
        "net/http"
        "os"
        "strconv"
        "zabbix_exporter/zabbix"
)

var gitTag string

var (
        listeningAddress = flag.String("telemetry.address", ":9051", "Address on which to expose metrics.")
        metricsEndpoint  = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
        zabbixAddr       = flag.String("zabbix_addr", "127.0.0.1", "Zabbix server addr")
        zabbixPort       = flag.Int("zabbix_port", 10051, "Zabbix server port")
        version          = flag.Bool("version", false, "Show version number and quit")
)

func main() {
        flag.Parse()
        if *version {
                fmt.Println(gitTag)
                os.Exit(0)
        }
        log.Printf("Scraping %s:%s", *zabbixAddr, strconv.Itoa(*zabbixPort))

        registerer := prometheus.DefaultRegisterer
        gatherer := prometheus.DefaultGatherer
        e := zabbix.NewZabbix(*zabbixAddr, *zabbixPort)
        registerer.MustRegister(e)

        http.Handle(*metricsEndpoint, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(`<html>
                        <head><title>Zabbix Exporter</title></head>
                        <body>
                        <h1>Zabbix Exporter</h1>
                        <p><a href="` + *metricsEndpoint + `">Metrics</a></p>
                        </body>
                        </html>`))
        })
        log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}