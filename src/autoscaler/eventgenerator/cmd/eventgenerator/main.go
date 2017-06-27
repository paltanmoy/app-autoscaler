package main

import (
	"autoscaler/db"
	"autoscaler/db/sqldb"
	"autoscaler/eventgenerator"
	"autoscaler/eventgenerator/aggregator"
	"autoscaler/eventgenerator/config"
	"autoscaler/eventgenerator/generator"
	"autoscaler/models"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/consuladapter"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.com/hashicorp/consul/api"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func main() {
	var path string
	flag.StringVar(&path, "c", "", "config file")
	flag.Parse()
	if path == "" {
		fmt.Fprintln(os.Stdout, "missing config file\nUsage:use '-c' option to specify the config file path")
		os.Exit(1)
	}

	conf, err := loadConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%s\n", err.Error())
		os.Exit(1)
	}

	logger := initLoggerFromConfig(&conf.Logging)
	egClock := clock.NewClock()

	appMetricDB, err := sqldb.NewAppMetricSQLDB(conf.DB.AppMetricDBUrl, logger.Session("appMetric-db"))
	if err != nil {
		logger.Error("failed to connect app-metric database", err, lager.Data{"url": conf.DB.AppMetricDBUrl})
		os.Exit(1)
	}
	defer appMetricDB.Close()

	var policyDB db.PolicyDB
	policyDB, err = sqldb.NewPolicySQLDB(conf.DB.PolicyDBUrl, logger.Session("policy-db"))
	if err != nil {
		logger.Error("failed to connect policy database", err, lager.Data{"url": conf.DB.PolicyDBUrl})
		os.Exit(1)
	}
	defer policyDB.Close()

	policyPoller := aggregator.NewPolicyPoller(logger, egClock, conf.Aggregator.PolicyPollerInterval, policyDB)

	triggersChan := make(chan []*models.Trigger, conf.Evaluator.TriggerArrayChannelSize)
	evaluators, err := createEvaluators(logger, conf, triggersChan, appMetricDB)
	if err != nil {
		logger.Error("failed to create Evaluators", err)
		os.Exit(1)
	}

	evaluationManager, err := generator.NewAppEvaluationManager(logger, conf.Evaluator.EvaluationManagerInterval, egClock,
		triggersChan, policyPoller.GetPolicies)
	if err != nil {
		logger.Error("failed to create Evaluation Manager", err)
		os.Exit(1)
	}

	appMonitorsChan := make(chan *models.AppMonitor, conf.Aggregator.AppMonitorChannelSize)
	metricPollers, err := createMetricPollers(logger, conf, appMonitorsChan, appMetricDB)
	aggregator, err := aggregator.NewAggregator(logger, egClock, conf.Aggregator.AggregatorExecuteInterval,
		appMonitorsChan, policyPoller.GetPolicies)
	if err != nil {
		logger.Error("failed to create Aggregator", err)
		os.Exit(1)
	}

	eventGenerator := ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		policyPoller.Start()

		for _, evaluator := range evaluators {
			evaluator.Start()
		}
		evaluationManager.Start()

		for _, metricPoller := range metricPollers {
			metricPoller.Start()
		}
		aggregator.Start()

		close(ready)

		<-signals
		aggregator.Stop()
		evaluationManager.Stop()
		policyPoller.Stop()

		return nil
	})

	members := grouper.Members{
		{"eventGenerator", eventGenerator},
	}

	if conf.Lock.ConsulClusterConfig != "" {
		consulClient, err := consuladapter.NewClientFromUrl(conf.Lock.ConsulClusterConfig)
		if err != nil {
			logger.Fatal("new consul client failed", err)
		}

		serviceClient := eventgenerator.NewServiceClient(consulClient, egClock)

		lockMaintainer := serviceClient.NewEventGeneratorLockRunner(
			logger,
			generateGUID(logger),
			conf.Lock.LockRetryInterval,
			conf.Lock.LockTTL,
		)
		registrationRunner := initializeRegistrationRunner(logger, consulClient, egClock)
		members = append(grouper.Members{{"lock-maintainer", lockMaintainer}}, members...)
		members = append(members, grouper.Member{"registration", registrationRunner})
	}

	var lockDB db.LockDB

	lockDB, err = sqldb.NewLockSQLDB(conf.DBLock.LockDBURL, logger.Session("lock-db"))
	if err != nil {
		logger.Error("failed to connect lock database", err, lager.Data{"url": conf.DBLock.LockDBURL})
		os.Exit(1)
	}
	defer lockDB.Close()

	dbLockMaintainer := ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
		ttl, err := strconv.Atoi(conf.DBLock.LockTTL)
		if err != nil {
			logger.Info("Failed to convert ttl")
		}
		lockTicker := time.NewTicker(time.Duration(ttl) * time.Second)
		acquireLockflag := true
		owner := getOwner(logger, conf)
		keytype := conf.DBLock.Type
		fmt.Println("******************************* Acquiring Lock ******************************")
		lock := models.Lock{Type: keytype, Owner: owner, Last_Modified_Timestamp: time.Now().Unix(), Ttl: ttl}
		isLockAcquired, lockErr := lockDB.AcquireLock(lock)
		if lockErr != nil {
			logger.Error("Lock Error", lockErr)
		}
		if isLockAcquired {
			fmt.Println("******************** I am ready with lock , relinquishing the ocntroil over the next process****************************")
			close(ready)
			acquireLockflag = false
			fmt.Println("************************* Lock Acquired on fisrt attempt?  =  ", isLockAcquired, "*****************************************")
		}
		fmt.Println("*************************** Inside Go routine ******************************")
		for {
			select {
			case <-signals:
				fmt.Println("************************** Lets quit  **********************")
				lockTicker.Stop()
				releaseErr := lockDB.ReleaseLock(owner)
				if releaseErr != nil {
					logger.Error("Lock Error", releaseErr)
				}
				acquireLockflag = true
				fmt.Println("*************************Releasing lock**************************")
				return nil

			case <-lockTicker.C:
				fmt.Println("********************************* retrying-acquiring-lock ************************************")
				lock := models.Lock{Type: keytype, Owner: owner, Last_Modified_Timestamp: time.Now().Unix(), Ttl: ttl}
				isLockAcquired, lockErr := lockDB.AcquireLock(lock)
				if lockErr != nil {
					logger.Error("Lock Error", lockErr)
				}
				fmt.Println("************************* Lock Acquired ?  =  ", isLockAcquired, "*****************************************")
				if isLockAcquired && acquireLockflag {
					fmt.Println("******************** I am ready with lock****************************")
					close(ready)
					acquireLockflag = false
				}
			}
		}
	})

	members = append(grouper.Members{{"db-lock-maintainer", dbLockMaintainer}}, members...)

	monitor := ifrit.Invoke(sigmon.New(grouper.NewOrdered(os.Interrupt, members)))

	logger.Info("started")

	err = <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func initLoggerFromConfig(conf *config.LoggingConfig) lager.Logger {
	logLevel, err := getLogLevel(conf.Level)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to initialize logger: %s\n", err.Error())
		os.Exit(1)
	}
	logger := lager.NewLogger("eventgenerator")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, logLevel))

	return logger
}

func getLogLevel(level string) (lager.LogLevel, error) {
	switch level {
	case "debug":
		return lager.DEBUG, nil
	case "info":
		return lager.INFO, nil
	case "error":
		return lager.ERROR, nil
	case "fatal":
		return lager.FATAL, nil
	default:
		return -1, fmt.Errorf("Error: unsupported log level:%s", level)
	}
}

func loadConfig(path string) (*config.Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %q: %s", path, err.Error())
	}

	configFileBytes, err := ioutil.ReadAll(configFile)
	configFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read data from config file %q: %s", path, err.Error())
	}

	conf, err := config.LoadConfig(configFileBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %s", path, err.Error())
	}

	err = conf.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %s", err.Error())
	}

	return conf, nil
}

func createEvaluators(logger lager.Logger, conf *config.Config, triggersChan chan []*models.Trigger, database db.AppMetricDB) ([]*generator.Evaluator, error) {
	count := conf.Evaluator.EvaluatorCount
	scalingEngineUrl := conf.ScalingEngine.ScalingEngineUrl

	tlsCerts := &conf.ScalingEngine.TLSClientCerts
	if tlsCerts.CertFile == "" || tlsCerts.KeyFile == "" {
		tlsCerts = nil
	}

	client := cfhttp.NewClient()
	client.Transport.(*http.Transport).MaxIdleConnsPerHost = count
	if tlsCerts != nil {
		tlsConfig, err := cfhttp.NewTLSConfig(tlsCerts.CertFile, tlsCerts.KeyFile, tlsCerts.CACertFile)
		if err != nil {
			return nil, err
		}
		client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	}

	evaluators := make([]*generator.Evaluator, count)
	for i := 0; i < count; i++ {
		evaluators[i] = generator.NewEvaluator(logger, client, scalingEngineUrl, triggersChan, database)
	}

	return evaluators, nil
}

func createMetricPollers(logger lager.Logger, conf *config.Config, appChan chan *models.AppMonitor, database db.AppMetricDB) ([]*aggregator.MetricPoller, error) {
	tlsCerts := &conf.MetricCollector.TLSClientCerts
	if tlsCerts.CertFile == "" || tlsCerts.KeyFile == "" {
		tlsCerts = nil
	}

	count := conf.Aggregator.MetricPollerCount
	client := cfhttp.NewClient()
	client.Transport.(*http.Transport).MaxIdleConnsPerHost = count
	if tlsCerts != nil {
		tlsConfig, err := cfhttp.NewTLSConfig(tlsCerts.CertFile, tlsCerts.KeyFile, tlsCerts.CACertFile)
		if err != nil {
			return nil, err
		}
		client.Transport.(*http.Transport).TLSClientConfig = tlsConfig
	}

	pollers := make([]*aggregator.MetricPoller, count)
	for i := 0; i < count; i++ {
		pollers[i] = aggregator.NewMetricPoller(logger, conf.MetricCollector.MetricCollectorUrl, appChan, client, database)
	}

	return pollers, nil
}

func initializeRegistrationRunner(
	logger lager.Logger,
	consulClient consuladapter.Client,
	clock clock.Clock) ifrit.Runner {
	registration := &api.AgentServiceRegistration{
		Name: "eventgenerator",
		Check: &api.AgentServiceCheck{
			TTL: "20s",
		},
	}
	return locket.NewRegistrationRunner(logger, registration, consulClient, locket.RetryInterval, clock)
}

func generateGUID(logger lager.Logger) string {
	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Fatal("Couldn't generate uuid", err)
	}
	return uuid.String()
}

func getOwner(logger lager.Logger, conf *config.Config) string {
	var owner string
	if strings.TrimSpace(os.Getenv("VCAP_APPLICATION")) != "" {
		logger.Info("Running on CF")
		appEnv, _ := cfenv.Current()
		fmt.Println("ID:", appEnv.ID)
		owner = appEnv.ID
	} else if conf.DBLock.Owner != "" {
		logger.Info("Running on VM throgh BOSH Release")
		owner = conf.DBLock.Owner
	} else {
		logger.Info("Running from binary")
		owner = generateGUID(logger)
	}
	return owner
}
