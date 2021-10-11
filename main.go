package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sschwartz96/hygieia-docker-collector/internal"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func main() {
	// Setup logger
	// logger, err := zap.NewProduction()
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error initializing logger:", err.Error())
	}

	// Catch run error
	err = run(logger)
	if err != nil {
		logger.Error("run() error", zap.Error(err))
	}
}

func run(logger *zap.Logger) error {
	// 1. Load config
	logger.Info("Loading Config")
	config, err := loadConfig()
	if err != nil {
		return err
	}
	logger.Info("Config Loaded")

	// 2. Connect to mongodb
	logger.Info("Connecting to mongo")
	mongoClient, err := connectMongo(config, logger)
	if err != nil {
		return err
	}
	mongoStore := internal.NewMongoStore(mongoClient.Database(config.MongoDBName), logger)
	logger.Info("Mongo Connected")

	// 3. Start the scheduler to begin collecting
	cronSchedule, err := cron.ParseStandard(config.Cron)
	if err != nil {
		return fmt.Errorf("Error parsing cron(%s): %v", config.Cron, err)
	}

	collector := internal.NewCollector(logger, mongoStore, config.CollectorName)
	nextTime := time.Now()
	for {
		nextTime = cronSchedule.Next(nextTime)
		logger.Info("Next Scheduled Collect", zap.Time("time", nextTime))
		time.Sleep(nextTime.Sub(time.Now()))

		ctx := context.TODO()
		err = collector.Collect(ctx)
		if err != nil {
			logger.Error("Error collecting", zap.Error(err))
		}
	}
}

func loadConfig() (*internal.Config, error) {
	configFileLocation := flag.String("config", "./config.json", "Set the config file")
	configFile, err := os.Open(*configFileLocation)
	if err != nil {
		return nil, fmt.Errorf("Error loading config file: %v", err)
	}
	config, err := internal.ParseConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v", err)
	}
	return config, nil
}

func connectMongo(config *internal.Config, logger *zap.Logger) (*mongo.Client, error) {
	mongoURI := fmt.Sprintf("mongodb://%s:%d/%s", config.MongoHost, config.MongoPort, config.MongoDBName)
	opts := options.Client().SetAppName("hygieia-docker-collector").ApplyURI(mongoURI)
	if len(config.MongoUser) > 0 && len(config.MongoPassword) > 0 {
		logger.Info("Setting MongoDB Auth", zap.String("username", config.MongoUser))
		opts = opts.SetAuth(options.Credential{Username: config.MongoUser, Password: config.MongoPassword})
	}

	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("Error creating Mongo client: %v", err)
	}
	ctx, cnlFn := context.WithTimeout(context.Background(), time.Second*10)
	defer cnlFn()
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to mongo db: %v", err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Error pinging the mongo db: %v", err)
	}
	return client, nil
}
