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
)

type Store interface {
	// CreateUniqueIndexes takes in a field to collection make to create unique indexes
	CreateUniqueIndexes(context.Context, map[string]string)

	// RegisterCollector registers collector to the database.
	RegisterCollector(context.Context) (*primitive.ObjectID, error)
	UpdateCollector(context.Context, *primitive.ObjectID) error

	FindAllCollectorItems(context.Context) ([]models.CollectorItem, error)
	UpdateCollectorItem(context.Context, *primitive.ObjectID) error

	UpsertContainers(context.Context, []types.Container)
	FindContainer(context.Context, string) (*models.Container, error)
	UpdateContainer(context.Context, *models.Container) error
	DeleteContainer(context.Context, string) error
}

type MongoStore struct {
	db     *mongo.Database
	logger *zap.Logger
}

func NewMongoStore(db *mongo.Database, logger *zap.Logger) *MongoStore {
	return &MongoStore{db: db, logger: logger}
}

func (m *MongoStore) CreateUniqueIndexes(ctx context.Context, fieldMap map[string]string) {
	for field, collection := range fieldMap {
		index := mongo.IndexModel{Keys: bson.M{field: 1}, Options: options.Index().SetUnique(true)}
		indexName, err := m.db.Collection(collection).Indexes().CreateOne(ctx, index)
		if err != nil {
			m.logger.Error("Error Creating Index", zap.Error(err))
		}
		m.logger.Debug("Created Index", zap.String("Name", indexName))
	}
}

// RegisterCollector registers collector to the database.
func (m *MongoStore) RegisterCollector(ctx context.Context) (*primitive.ObjectID, error) {
	c := models.Collector{
		Name:                     "docker-hygieia-collector",
		CollectorType:            "Docker",
		Enabled:                  true,
		Online:                   true,
		Errors:                   nil,
		UniqueFields:             nil,
		AllFields:                nil,
		LastExecuted:             time.Now().Unix(),
		LastExectuedTime:         time.Now(),
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

func (m *MongoStore) UpdateCollector(ctx context.Context, id *primitive.ObjectID) error {
	res, err := m.db.Collection(collectorCollection).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"lastExecuted":     time.Now().Unix(),
			"lastExecutedTime": time.Now(),
		}},
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
	res, err := m.db.Collection(collectorItemsCollection).UpdateByID(ctx, id, bson.M{
		"lastUpdated": time.Now().Unix(),
	})
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

func (m *MongoStore) FindContainer(_ context.Context, _ string) (*models.Container, error) {
	panic("not implemented") // TODO: Implement
}

func (m *MongoStore) UpdateContainer(_ context.Context, _ *models.Container) error {
	panic("not implemented") // TODO: Implement
}

func (m *MongoStore) DeleteContainer(_ context.Context, _ string) error {
	panic("not implemented") // TODO: Implement
}
