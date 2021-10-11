package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/sschwartz96/hygieia-docker-collector/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const (
	collectorCollection      = "collectors"
	collectorItemsCollection = "collector_items"
	containerCollection      = "docker_containers"
	containerStatsCollection = "docker_container_stats"
	networkCollection        = "docker_networks"
	volumeCollection         = "docker_volumes"
)

type Store interface {
	// CreateUniqueIndexes takes in a field to collection make to create unique indexes
	CreateUniqueIndexes(context.Context, []UniqueIndex)

	DoesCollectorExist(context.Context, string) (*primitive.ObjectID, error)
	// RegisterCollector registers collector to the database.
	RegisterCollector(context.Context, string) (*primitive.ObjectID, error)
	// UpdateCollector takes in the collector id and the duration of the last collection
	UpdateCollector(context.Context, *primitive.ObjectID, time.Duration) error

	FindAllCollectorItems(context.Context) ([]models.CollectorItem, error)
	UpdateCollectorItem(context.Context, *primitive.ObjectID) error

	UpsertContainers(context.Context, []types.Container)

	UpsertNetworks(context.Context, []types.NetworkResource)

	UpsertVolumes(context.Context, []*types.Volume)

	UpsertContiainerStats(context.Context, types.StatsJSON) error
}

type MongoStore struct {
	db     *mongo.Database
	logger *zap.Logger
}

type UniqueIndex struct {
	field      string
	collection string
}

func NewMongoStore(db *mongo.Database, logger *zap.Logger) *MongoStore {
	return &MongoStore{db: db, logger: logger}
}

func (m *MongoStore) CreateUniqueIndexes(ctx context.Context, indexes []UniqueIndex) {
	for _, index := range indexes {
		indexModel := mongo.IndexModel{Keys: bson.M{index.field: 1}, Options: options.Index().SetUnique(true)}
		indexName, err := m.db.Collection(index.collection).Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			m.logger.Error("Error Creating Index", zap.Error(err))
		}
		m.logger.Debug("Created Index", zap.String("Name", indexName))
	}
}

func (m *MongoStore) DoesCollectorExist(ctx context.Context, name string) (*primitive.ObjectID, error) {
	res := m.db.Collection(collectorCollection).FindOne(ctx, bson.M{"name": name})
	err := res.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			m.logger.Info("ErrNoDocuments)")
			return nil, nil
		}
		return nil, err
	}
	var collector models.Collector
	err = res.Decode(&collector)
	if err != nil {
		return nil, err
	}
	return &collector.ID, err
}

// RegisterCollector registers collector to the database.
func (m *MongoStore) RegisterCollector(ctx context.Context, name string) (*primitive.ObjectID, error) {
	c := models.Collector{
		ID:                       primitive.NewObjectID(),
		Name:                     name,
		CollectorType:            "Docker",
		Enabled:                  true,
		Online:                   true,
		Errors:                   nil,
		UniqueFields:             nil,
		AllFields:                nil,
		LastExecuted:             time.Now().Unix(),
		LastExecutedTime:         time.Now(),
		LastExecutionRecordCount: 0,
		LastExecutedSeconds:      0,
	}
	res, err := m.db.Collection(collectorCollection).InsertOne(ctx, &c)
	if err != nil {
		return nil, err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("RegisterCollector() error: invalid inserted id")
	}
	return &id, err
}

func (m *MongoStore) UpdateCollector(ctx context.Context, id *primitive.ObjectID, lastExecutedDuration time.Duration) error {
	res, err := m.db.Collection(collectorCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"lastExecuted":        time.Now().Unix(),
				"lastExecutedTime":    time.Now(),
				"lastExecutedSeconds": lastExecutedDuration.Seconds(),
			},
			"$inc": bson.M{"lastExecutionRecordCount": 1},
		},
	)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("UpdateCollector() error. ModifiedCount == 0")
	}
	return nil
}

func (m *MongoStore) FindAllCollectorItems(ctx context.Context) ([]models.CollectorItem, error) {
	cur, err := m.db.Collection(collectorItemsCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var items []models.CollectorItem
	err = cur.All(ctx, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateCollectorItem takes collector_item id and updates the lastUpdated time
func (m *MongoStore) UpdateCollectorItem(ctx context.Context, id *primitive.ObjectID) error {
	res, err := m.db.Collection(collectorItemsCollection).UpdateByID(ctx, id,
		bson.M{"$set": bson.M{"lastUpdated": time.Now().Unix()}},
	)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("UpdateCollectorItem() error: modified count for id=%s", id.Hex())
	}
	return nil
}

func (m *MongoStore) UpsertContainers(ctx context.Context, containers []types.Container) {
	for _, c := range containers {
		_, err := m.db.Collection(containerCollection).UpdateOne(
			ctx,
			bson.M{"id": c.ID},
			bson.M{"$set": &c},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			m.logger.Error("UpsertContainers() error upserting", zap.Error(err), zap.String("ID", c.ID))
		}
	}
}

func (m *MongoStore) UpsertNetworks(ctx context.Context, networks []types.NetworkResource) {
	for _, n := range networks {
		_, err := m.db.Collection(networkCollection).UpdateOne(
			ctx, bson.M{"id": n.ID},
			bson.M{"$set": &n},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			m.logger.Error("UpsertNetworks() error upserting", zap.Error(err), zap.String("ID", n.ID))
		}
	}
}

func (m *MongoStore) UpsertVolumes(ctx context.Context, volumes []*types.Volume) {
	for _, v := range volumes {
		_, err := m.db.Collection(volumeCollection).UpdateOne(
			ctx, bson.M{"name": v.Name},
			bson.M{"$set": v},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			m.logger.Error("UpsertVolumes() error upserting", zap.Error(err), zap.String("Name", v.Name))
		}
	}
}

func (m *MongoStore) UpsertContiainerStats(ctx context.Context, stats types.StatsJSON) error {
	_, err := m.db.Collection(containerStatsCollection).UpdateOne(
		ctx, bson.M{"id": stats.ID},
		bson.M{"$set": &stats},
		options.Update().SetUpsert(true),
	)
	return err
}
