package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"time"

	"github.com/chengchung/ServerStatus/common/logger"
	"github.com/chengchung/ServerStatus/common/timer"
	"github.com/chengchung/ServerStatus/data/promql"
	"github.com/chengchung/ServerStatus/datasource"
	"github.com/chengchung/ServerStatus/proto"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/sirupsen/logrus"
)

var (
	iris_app  *iris.Application
	apiclient *promql.QueryClient

	config_refresh_interval time.Duration
	metric_scrape_interval  time.Duration

	scrape_timer *timer.DowngradeTimer

	listen_addr string

	configFlag = flag.String("config", "./conf.json", "path to config file")
)

func stats(ctx iris.Context) {
	scrape_timer.Trigger()
	report := apiclient.GetReport()
	ctx.StatusCode(200)
	ctx.JSON(report)
}

func reload(ctx iris.Context) {
	reload_conf()
	ctx.StatusCode(200)
}

func init_logger(path string) {
	logfmt := &logger.CustomFormatter{
		EnableGoRoutineId: true,
	}
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(logfmt)
	if len(path) == 0 {
		path = "./server.log"
	} else {
		path = path + "/server.log"
	}
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		logrus.SetOutput(logFile)
	}
}

func init_iris(path string) {
	app := iris.New()
	app.Logger().SetLevel("debug")
	app.Logger().Debugf(`Log level set to "debug"`)
	// Register the accesslog middleware.
	if len(path) == 0 {
		path = "./access.log"
	} else {
		path = path + "/access.log"
	}
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		// Close the file on shutdown.
		app.ConfigureHost(func(su *iris.Supervisor) {
			su.RegisterOnShutdown(func() {
				logFile.Close()
			})
		})

		ac := accesslog.New(logFile)
		ac.AddOutput(app.Logger().Printer)
		app.UseRouter(ac.Handler)
		app.Logger().Debugf("Using <%s> to log requests", logFile.Name())
	}

	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	})
	appParty := app.Party("/", crs).AllowMethods(iris.MethodOptions)
	appParty.Get("/json/stats.json", stats)
	appParty.Get("/reload", reload)

	iris_app = app
}

func read_conf() (*proto.MainConf, bool) {
	jsonFile, err := os.Open(*configFlag)
	if err != nil {
		logrus.Error(err)
		return nil, false
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var mc proto.MainConf
	if err = json.Unmarshal(byteValue, &mc); err != nil {
		logrus.Error(err)
		return nil, false
	}

	return &mc, true
}

func init_ds(cfg []interface{}) bool {
	if err := datasource.InitDS(cfg); err != nil {
		logrus.Error(err)
		return false
	}

	return true
}

func init_main() bool {
	flag.Parse()
	cfg, succ := read_conf()
	if !succ {
		logrus.Error("fail to read conf file")
		return false
	}

	listen_addr = cfg.Listen

	init_logger(cfg.LogPath)
	init_iris(cfg.LogPath)

	if cfg.RefreshInterval > 0 {
		config_refresh_interval = time.Duration(time.Second * time.Duration(cfg.RefreshInterval))
	} else {
		config_refresh_interval = time.Duration(time.Minute)
	}

	if cfg.ScrapeInterval > 0 {
		metric_scrape_interval = time.Duration(time.Second * time.Duration(cfg.ScrapeInterval))
	} else {
		metric_scrape_interval = time.Duration(time.Second * 15)
	}

	ds_cfg := make([]interface{}, 0)
	for _, ds := range cfg.DataSources {
		ds_cfg = append(ds_cfg, ds)
	}
	if !init_ds(ds_cfg) {
		return false
	}

	apiclient = promql.NewAPIClient(cfg.Nodes)

	return true
}

func reload_conf() {
	cfg, succ := read_conf()
	if !succ {
		logrus.Error("fail to read conf file")
		return
	}

	apiclient.ResetConf(cfg.Nodes)
}

func main() {
	if !init_main() {
		os.Exit(1)
	}

	go func() {
		for {
			t := time.NewTimer(config_refresh_interval)
			<-t.C
			reload_conf()
		}
	}()

	go func() {
		scrape_timer = timer.NewDowngradeTimer(metric_scrape_interval, metric_scrape_interval*128, 5)
		for {
			logrus.Info("start metric update")
			apiclient.Refresh()
			scrape_timer.Wait()
		}
	}()

	if len(listen_addr) == 0 {
		iris_app.Run(iris.Addr("127.0.0.1:39999"))
	} else {
		iris_app.Run(iris.Addr(listen_addr))
	}
}
