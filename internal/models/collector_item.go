package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectorType string

// Collector model used to register the collector in the db
type Collector struct {
	Name                     string                 `bson:"name"`
	CollectorType            CollectorType          `bson:"collectorType"`
	Enabled                  bool                   `bson:"enabled"`
	Online                   bool                   `bson:"online"`
	Errors                   []CollectionError      `bson:"errors"`
	UniqueFields             map[string]interface{} `bson:"uniqueFields"`
	AllFields                map[string]interface{} `bson:"allFields"`
	LastExecuted             int64                  `bson:"lastExecuted"`
	LastExectuedTime         time.Time              `bson:"lastExectuedTime"`
	LastExecutionRecordCount int64                  `bson:"lastExecutionRecordCount"`
	LastExecutedSeconds      int64                  `bson:"lastExecutedSeconds"`
}

type CollectorItem struct {
	ID          primitive.ObjectID `bson:"_id"'`
	Description string             `bson:"description"`
	NiceName    string             `bson:"niceName"`
	Environment string             `bson:"environment"`
	Enabled     bool               `bson:"enabled"`
	Errors      []CollectionError  `bson:"errors"`
	Pushed      bool               `bson:"pushed"`
	CollectorId primitive.ObjectID `bson:"collectorId"`
	LastUpdated int64              `bson:"lastUpdated"`
	// Expecting: { "host": "localhost", "apiVersion": "1.40", "port": "1234"}
	Options       map[string]string `bson:"options"`
	RefreshLink   string            `bson:"refreshLink"`
	AltIdentifier string            `bson:"altIdentifier"`
}

type CollectionError struct {
	ErrorCode     string `bson:"errorCode"`
	ErrorMessages string `bson:"errorMessages"`
	Timestamp     int64  `bson:"timestamp"`
}
