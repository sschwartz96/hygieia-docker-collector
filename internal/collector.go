package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sschwartz96/hygieia-docker-collector/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Collector struct {
	logger  *zap.Logger
	store   Store
	name    string
	mongoID *primitive.ObjectID
	once    *sync.Once
}

func NewCollector(logger *zap.Logger, store Store, name string) *Collector {
	return &Collector{logger: logger, once: &sync.Once{}, store: store, name: name}
}

func (c *Collector) Collect(ctx context.Context) error {
	startTime := time.Now()
	// Need to register collector on first run
	c.once.Do(func() {
		c.register(ctx)
		c.store.CreateUniqueIndexes(ctx,
			[]UniqueIndex{
				{"id", containerCollection},
				{"id", networkCollection},
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
			// update collector item in db
			err = c.store.UpdateCollectorItem(ctx, &item.ID)
			if err != nil {
				c.logger.Error("Error updating collector item", zap.Error(err), zap.String("ID", item.ID.Hex()))
			}
		}
	}

	c.logger.Info("Finished collecting...")

	elapsed := time.Now().Sub(startTime)
	err = c.store.UpdateCollector(ctx, c.mongoID, elapsed)
	if err != nil {
		c.logger.Error("Error Updating Collector in DB", zap.Error(err))
	}
	return nil
}

func (c *Collector) register(ctx context.Context) {
	// check if the the collector exists based on the name
	id, err := c.store.DoesCollectorExist(ctx, c.name)
	if err != nil {
		c.logger.Fatal("Error ", zap.Error(err))
		// os.Exit(1)
	}
	if id != nil {
		c.mongoID = id
		return
	}
	// register collector
	id, err = c.store.RegisterCollector(ctx, c.name)
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

	// docker client collect containers
	containers, err := client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		c.logger.Error("collectItem() error collecting containers",
			zap.String("host", item.Options["host"]),
			zap.Error(err),
		)
	}
	c.store.UpsertContainers(ctx, containers)

	// get the stats for the containers
	for i := range containers {
		stats, err := client.ContainerStats(ctx, containers[i].ID, false)
		if err != nil {
			c.logger.Error("collectItem() error collecting container stats",
				zap.String("Host", item.Options["host"]),
				zap.String("Container ID", containers[i].ID),
				zap.Error(err),
			)
		} else {
			statDecoder := json.NewDecoder(stats.Body)
			var containerStat types.StatsJSON
			err := statDecoder.Decode(&containerStat)
			if err != nil {
				c.logger.Error("collectItem() error collecting container stats",
					zap.Error(err),
					zap.String("container_id", containers[i].ID),
				)
			}
			err = stats.Body.Close()
			if err != nil {
				c.logger.Error("collectItem() error closing stats body io.Reader",
					zap.Error(err),
					zap.String("container_name", containerStat.Name),
					zap.String("container_id", containerStat.ID),
				)
			}
			err = c.store.UpsertContiainerStats(ctx, containerStat)
			if err != nil {
				c.logger.Error("collectItem() error upserting container stats",
					zap.Error(err),
					zap.String("container_name", containerStat.Name),
					zap.String("container_id", containerStat.ID),
				)
			}
		}
	}

	// get the docker networks
	networks, err := client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		c.logger.Error("collectItem() error collecting networks",
			zap.String("Host", item.Options["host"]),
			zap.Error(err),
		)
	} else {
		c.store.UpsertNetworks(ctx, networks)
	}

	// get the docker volumes
	volumes, err := client.VolumeList(ctx, filters.NewArgs())
	if err != nil {
		c.logger.Error("collectItem() error collecting volumes",
			zap.String("Host", item.Options["host"]),
			zap.Error(err),
		)
	} else {
		if len(volumes.Warnings) > 0 {
			c.logger.Info("collectItem() warnings when collecting volumes",
				zap.Array("warnings",
					zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
						for i := range volumes.Warnings {
							ae.AppendString(volumes.Warnings[i])
						}
						return nil
					}),
				),
			)
		}
		c.store.UpsertVolumes(ctx, volumes.Volumes)
	}
	// TODO: collect other docker info
	return nil
}
