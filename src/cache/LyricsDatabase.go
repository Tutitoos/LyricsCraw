package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// LyricsDatabase is a small MongoDB-backed cache implementation.
type LyricsDatabase struct {
	client     *mongo.Client
	collection *mongo.Collection
	ttl        time.Duration
}

// LyricsDB is the global DB-backed cache instance (nil until initialized).
var LyricsDB *LyricsDatabase

// cacheDoc is the BSON representation stored in MongoDB.
type cacheDoc struct {
	Key       string    `bson:"_id"`
	Value     string    `bson:"value"`
	ExpiresAt time.Time `bson:"expires_at"`
}

// InitLyricsDBFromEnv initializes LyricsDB using MONGODB_URI and APP_LYRICS_CACHE_TTL_SECONDS.
// Environment variables:
// - MONGODB_URI (required)
// - MONGODB_DB (optional, default: "lyricscrawl")
// - MONGODB_COLLECTION (optional, default: "lyrics_cache")
func InitLyricsDBFromEnv() error {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return fmt.Errorf("MONGODB_URI not set")
	}
	ttl := 1800 * time.Second
	if v := os.Getenv("APP_LYRICS_CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Second
		}
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "lyricscrawl"
	}
	collName := os.Getenv("MONGODB_COLLECTION")
	if collName == "" {
		collName = "lyrics_cache"
	}

	d := &LyricsDatabase{ttl: ttl}
	if err := d.Connect(uri, dbName, collName); err != nil {
		return err
	}
	LyricsDB = d
	return nil
}

// Connect establishes a MongoDB client, ensures the TTL index exists on expires_at.
func (d *LyricsDatabase) Connect(uri, dbName, collName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}
	// verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return err
	}
	coll := client.Database(dbName).Collection(collName)

	// create TTL index on expires_at (expireAfterSeconds: 0)
	idxModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}
	if _, err := coll.Indexes().CreateOne(ctx, idxModel); err != nil {
		// not fatal: continue but return error
		_ = client.Disconnect(ctx)
		return fmt.Errorf("failed to ensure TTL index: %w", err)
	}

	d.client = client
	d.collection = coll

	fmt.Printf("✅ Connected to MongoDB, using DB: %s, Collection: %s, TTL: %v\n", dbName, collName, d.ttl)

	return nil
}

// Get returns the cached value if present and not expired.
func (d *LyricsDatabase) Get(key string) (string, bool) {
	if d == nil || d.collection == nil {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var doc cacheDoc
	err := d.collection.FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if err != nil {
		return "", false
	}
	if time.Now().After(doc.ExpiresAt) {
		// expired: remove and treat as miss
		go func(k string) {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel2()
			d.collection.DeleteOne(ctx2, bson.M{"_id": k})
		}(key)
		return "", false
	}
	return doc.Value, true
}

// Set writes the value with TTL (upsert).
func (d *LyricsDatabase) Set(key, value string) error {
	if d == nil || d.collection == nil {
		return errors.New("db not initialized")
	}
	expires := time.Now().Add(d.ttl)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	doc := cacheDoc{Key: key, Value: value, ExpiresAt: expires}
	_, err := d.collection.UpdateOne(ctx, bson.M{"_id": key}, bson.M{"$set": doc}, options.Update().SetUpsert(true))
	return err
}

// Close closes the MongoDB client connection.
func (d *LyricsDatabase) Close() {
	if d == nil || d.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = d.client.Disconnect(ctx)
	d.client = nil
	d.collection = nil
}
