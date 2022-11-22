package db

import (
	"context"
	"log"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client mongo.Client

func init() {
	if viper.GetBool("database") {
		uri := viper.GetString("mongodb_uri")
		if uri == "" {
			log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
		}
		client, err := mongo.NewClient(options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatal(err)
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)
	}
}

func Count(collection string, filter bson.D) (error, int64) {
	coll := client.Database("opsilon").Collection(collection)
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return err, 0
	}
	return nil, count
}

func InsertOne(collection string, filter bson.D, doc interface{}) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		return err
	}
	return nil
}

func InsertMany(collection string, filter bson.D, docs []interface{}) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.InsertMany(context.TODO(), docs)
	if err != nil {
		return err
	}
	return nil
}

func UpdateOne(collection string, filter bson.D, update bson.D) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMany(collection string, filter bson.D, update bson.D) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func ReplaceOne(collection string, filter bson.D, replacement bson.D) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.ReplaceOne(context.TODO(), filter, replacement)
	if err != nil {
		return err
	}
	return nil
}

func DeleteOne(collection string, filter bson.D) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMany(collection string, filter bson.D) error {
	coll := client.Database("opsilon").Collection(collection)
	_, err := coll.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func FindOne(collection string, filter bson.D, doc interface{}) error {
	coll := client.Database("opsilon").Collection(collection)
	// filter := bson.D{{"name", "Bagels N Buns"}}
	err := coll.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return err
		}
		return err
	}
	return nil
}

func FindMany(collection string, filter bson.D, docs []interface{}) error {
	coll := client.Database("opsilon").Collection(collection)
	// filter := bson.D{{"name", "Bagels N Buns"}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	if err = cursor.All(context.TODO(), &docs); err != nil {
		return err
	}

	for _, doc := range docs {
		cursor.Decode(&doc)
		if err != nil {
			return err
		}
	}
	return err
}

// func InsertManyUpsert(ctx context.Context, col string, filter bson.D, docs []interface{}) error {
// 	opsilonDatabase := client.Database("opsilon")
// 	repo := opsilonDatabase.Collection(col)
// 	insertResult, err := repo.UpdateMany(ctx, filter, docs)
// 	update := bson.D{{"$set", bson.D{{"age", 1}}}}

// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("Inserted %v documents into %s collection!\n", len(insertResult.InsertedIDs), col)
// 	return nil
// }
