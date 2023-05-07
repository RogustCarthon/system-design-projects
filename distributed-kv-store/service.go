package distributed_kv_store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connect() (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo. [%w]", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping. [%w]", err)
	}
	log.Info().Msg("connected to mongo")

	return client.Database("db"), nil
}

func Init() Service {
	db, err := connect()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect")
	}

	return Service{col: db.Collection("c")}
}

func (s *Service) Cleanup(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("batch cleanup exiting")
			return
		case <-t.C:
			if res, err := s.col.DeleteMany(ctx, bson.M{
				"expiry": bson.M{"$lte": time.Now().Unix()},
			}); err == nil {
				log.Info().Int64("cnt", res.DeletedCount).Msg("deleted records")
			} else {
				log.Error().Err(err).Msg("failed to delete")
			}
		}
	}
}

type Service struct {
	col *mongo.Collection
}

type KV struct {
	Key    string      `bson:"key"`
	Value  interface{} `bson:"value"`
	Expiry int64       `bson:"expiry"`
}

type KVInput struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   int64       `json:"TTL"`
}

type KVOutput struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   int64       `json:"TTL"`
}

func (s *Service) Get(ctx context.Context, k string) (*KVOutput, error) {
	if res := s.col.FindOne(ctx, bson.M{
		"key":    k,
		"expiry": bson.M{"$gte": time.Now().Unix()},
	}); res.Err() == nil {
		output := &KVOutput{}
		if err := res.Decode(output); err != nil {
			return nil, fmt.Errorf("failed to decode. [%w]", err)
		}
		return output, nil
	} else {
		return nil, fmt.Errorf("failed to find. [%w]", res.Err())
	}
}

func (s *Service) GetCount(ctx context.Context) (int64, error) {
	c, err := s.col.CountDocuments(ctx, bson.M{
		"expiry": bson.M{"$gte": time.Now().Unix()}},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get count. [%w]", err)
	}
	return c, nil
}

func (s *Service) Put(ctx context.Context, kv KVInput) error {
	if _, err := s.col.UpdateOne(
		ctx, bson.M{"key": kv.Key},
		bson.M{"$set": KV{
			Key:    kv.Key,
			Value:  kv.Value,
			Expiry: time.Now().Unix() + kv.TTL,
		}},
		options.Update().SetUpsert(true),
	); err != nil {
		return fmt.Errorf("failed to insert. [%w]", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, k string) error {
	if _, err := s.col.UpdateOne(
		ctx, bson.M{"key": k},
		bson.M{"$set": bson.M{"expiry": -1}},
	); err != nil {
		return fmt.Errorf("failed to update. [%w]", err)
	}
	return nil
}
