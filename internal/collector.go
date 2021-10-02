package internal

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sschwartz96/hygieia-docker-collector/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type Collector struct {
	logger  *zap.Logger
	store   Store
	mongoID *primitive.ObjectID
	once    *sync.Once
}

func NewCollector(logger *zap.Logger, store Store) *Collector {
	return &Collector{logger: logger, once: &sync.Once{}, store: store}
}

func (c *Collector) Collect(ctx context.Context) error {
	// Need to register collector on first run
	c.once.Do(func() {
		c.register(ctx)
		c.store.CreateUniqueIndexes(ctx,
			map[string]string{
				"id": containerCollection,
			},
		)
	})

	c.logger.Info("Starting to collect...")

	collectorItems, err := c.store.FindAllCollectorItems(ctx)
	if err != nil {
		return err
	}
	c.logger.Debug("collectorItems", zap.Int("Length", len(collectorItems)))

	for _, item := range collectorItems {
		if item.Enabled {
			err = c.collectItem(ctx, item)
			if err != nil {
				c.logger.Error("Error collecting item", zap.Error(err))
			}
		}
	}

	c.logger.Info("Finished collecting...")

	err = c.store.UpdateCollector(ctx, c.mongoID)
	if err != nil {
		c.logger.Error("Error Updating Collector in DB", zap.Error(err))
	}
	return nil
}

func (c *Collector) register(ctx context.Context) {
	id, err := c.store.RegisterCollector(ctx)
	if err != nil {
		c.logger.Fatal("Error Registering Collector", zap.Error(err))
		os.Exit(1)
	}
	c.mongoID = id
}

func (c *Collector) collectItem(ctx context.Context, item models.CollectorItem) error {
	host, ok := item.Options["host"]
	if !ok {
		return fmt.Errorf("CollectorItem: id: %s has no host", item.ID.Hex())
	}
	// port, ok := item.Options["port"]
	// if !ok {
	// 	return fmt.Errorf("CollectorItem: id: %s has no port", item.ID.Hex())
	// }
	version, ok := item.Options["apiVersion"]
	if !ok {
		return fmt.Errorf("CollectorItem: id: %s has no version", item.ID.Hex())
	}
	client, err := client.NewClientWithOpts(client.WithHost(host), client.WithVersion(version))
	if err != nil {
		return err
	}
	ping, err := client.Ping(ctx)
	if err != nil {
		return err
	}
	c.logger.Debug("Docker Client Ping", zap.String("APIVersion", ping.APIVersion), zap.String("OSType", ping.OSType))

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{})
	c.store.UpsertContainers(ctx, containers)

	return nil
}
