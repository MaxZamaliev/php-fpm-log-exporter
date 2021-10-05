// php-fpm_exporter - exports metrics for Prometheus from php-fpm status page
package main

import (
	"exporterHTTPServer"
	"flag"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "0.1"

// command line options default values
var listenAddress = "*"
var listenPort = 9253 // https://github.com/prometheus/prometheus/wiki/Default-port-allocations
var metricsPath = "/metrics"
var logFile = "/var/log/php-fpm/www-access.log"
var debugParse = false

// init logging
var debugLog = log.New(os.Stdout, "php-fpm_exporter: DEBUG\t", log.Ldate|log.Ltime|log.Lmsgprefix)
var infoLog = log.New(os.Stdout, "php-fpm_exporter: INFO\t", log.Ldate|log.Ltime|log.Lmsgprefix)
var errorLog = log.New(os.Stderr, "php-fpm_exporter: ERROR\t", log.Ldate|log.Ltime|log.Lmsgprefix)


func init() {
	infoLog.Println("Starting php-fpm-log-exporter version " + version)

	// get command line options
	flag.StringVar(&listenAddress, "listen-address", listenAddress, "ip-address where exporter listens connectioins from Prometheus '<ip-addr>|localhost|*|any'")
	flag.IntVar(&listenPort, "listen-port", listenPort, "port where exporter listens connections from Prometheus '1-65535'")
	flag.StringVar(&metricsPath, "metrics-path", metricsPath, "path after http://<listen-address>:[<listen-port]/ from where exporter returns metrics")
	flag.StringVar(&logFile, "log-file", logFile, "path and file to access log of php-fpm")
	flag.BoolVar(&debugParse, "debug-parse", debugParse, "enable verbosity output for parsing")
	flag.Parse()

	infoLog.Printf("listens on http://%s:%v%s", listenAddress, listenPort, metricsPath)
	infoLog.Printf("parsing log file: %s",logFile)

	if debugParse {
		debugLog.Println("debug-parse enabled")
	} else {
		debugLog.Println("debug-parse disabled")
	}
}

func main() {

        errorChan := make(chan error)
        srv := &exporterHTTPServer.Server {
            ListenAddress: listenAddress,
            ListenPort: listenPort,
            MetricsPath: metricsPath,
            ErrorChan: errorChan,
            Handler: promhttp.Handler(),
        }

        srv.Start()
        go parse(logFile, errorChan)

        for err := range errorChan {
            errorLog.Println(err)
        }

        infoLog.Println("Stop php-fpm-log-exporter")
        os.Exit(0)
}

func parse(fileName string, errorChan chan error) error {
        // start tailf log file
        t,err := tail.TailFile(fileName,tail.Config{
            Follow:true,
            MustExist: true,
            ReOpen: true,
        })
        if err !=nil {
            errorChan<-err
            close(errorChan)
            return err
        }

        // init metrics variables
        requestsCPU := prometheus.NewSummary(
            prometheus.SummaryOpts {
                Namespace: "phpfpm",
                Subsystem: "requests",
                Name: "cpu",
                Help: "Summary requests cpu use.",
        })
        prometheus.MustRegister(requestsCPU)

        requestsMEM := prometheus.NewSummary(
            prometheus.SummaryOpts {
                Namespace: "phpfpm",
                Subsystem: "requests",
                Name: "memory",
                Help: "Summary requests memory use.",
        })
        prometheus.MustRegister(requestsMEM)

        requestsDUR := prometheus.NewSummary(
            prometheus.SummaryOpts {
                Namespace: "phpfpm",
                Subsystem: "requests",
                Name: "duration",
                Help: "Summary requests duration.",
        })
        prometheus.MustRegister(requestsDUR)

        httpREQ :=prometheus.NewCounterVec(
            prometheus.CounterOpts {
                Namespace: "phpfpm",
                Subsystem: "requests",
                Name: "total",
                Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
            },
            []string{"code","method"},
        )
        prometheus.MustRegister(httpREQ)

        // parsing lines from log file
        for line := range t.Lines {
            l := strings.Replace(line.Text,"  "," ",-1)
            l = strings.Replace(l,"  "," ",-1)
            l = strings.Replace(l,"\"","",-1)

            words := strings.Split(l," ")
            if len(words) != 8 {
        	if debugParse {
                    debugLog.Printf("can't parse string from log: %v",line.Text)
                }
                continue
            }

            cpu := words[2]
            if m, _ := regexp.MatchString(`^\d*\.\d*%$`, cpu); !m {
        	if debugParse {
                    debugLog.Printf("can't parse CPU from log string: %v",line.Text)
                }
                continue
            }
            cpu = strings.Replace(words[2],"%","",-1)
            fcpu,_ := strconv.ParseFloat(cpu,64)

            mem := words[3]
            if m, _ := regexp.MatchString(`^\d*$`, mem); !m {
        	if debugParse {
                    debugLog.Printf("can't parse MEMORY from log string: %v",line.Text)
                }
                continue
            }
            fmem,_ := strconv.ParseFloat(mem,64)

            dur := words[4]
            if m, _ := regexp.MatchString(`^\d*\.\d*$`, dur); !m {
        	if debugParse {
                    debugLog.Printf("can't parse DURATION from log string: %v",line.Text)
                }
                continue
            }
            fdur,_ := strconv.ParseFloat(dur,64)

            code := words[5]
            if m, _ := regexp.MatchString(`^[2345]\d\d$`, code); !m {
        	if debugParse {
                    debugLog.Printf("can't parse CODE from log string: %v",line.Text)
                }
                continue
            }

            method := words[6]
            if method != "GET" && method != "POST" {
        	if debugParse {
                    debugLog.Printf("can't parse METHOD from log string: %v",line.Text)
                }
                continue
            }

            requestsCPU.Observe(fcpu)
            requestsMEM.Observe(fmem)
            requestsDUR.Observe(fdur)
            httpREQ.WithLabelValues(code,method).Inc()

    	    if debugParse {
                debugLog.Printf("Parsed string from log: cpu=%f; memory=%f; duration=%f code=%s method=%s",fcpu, fmem, fdur, code, method)
            }
    }
    return nil
}