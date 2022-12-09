package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client mongo.Client

func Init() {
	dbEnabled := viper.GetBool("database")
	fmt.Println("DB Enabled:", dbEnabled)
	if viper.GetBool("database") {
		fmt.Println("DB Enabled")
		uri := viper.GetString("mongodb_uri")
		fmt.Println("MongoDB URI", uri)
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
		client.Database("opsilon")
		defer client.Disconnect(ctx)
	}
}

func WebSocket(ws *websocket.Conn) {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection("logs")
	logStream, err := coll.Watch(context.TODO(), mongo.Pipeline{})
	if err != nil {
		panic(err)
	}

	defer logStream.Close(context.TODO())

	for logStream.Next(context.TODO()) {
		var data bson.M
		if err := logStream.Decode(&data); err != nil {
			panic(err)
		}
		fmt.Println("data", data)
		err = ws.WriteJSON(data)
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}
}

func Count(collection string, filter bson.D) (error, int64) {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	count, err := coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return err, 0
	}
	return nil, count
}

func InsertOne(collection string, doc interface{}) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.InsertOne(context.TODO(), doc)
	if err != nil {
		return err
	}
	return nil
}

func InsertMany(collection string, docs []interface{}) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.InsertMany(context.TODO(), docs)
	if err != nil {
		return err
	}
	return nil
}

func UpdateOne(collection string, filter bson.D, update interface{}) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.UpdateOne(context.TODO(), filter, update, &options.UpdateOptions{Upsert: mongo.NewUpdateOneModel().Upsert})
	if err != nil {
		return err
	}
	return nil
}

func UpdateByID(collection string, id string, update interface{}) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.UpdateByID(context.TODO(), bson.D{{"_id", id}}, update, &options.UpdateOptions{Upsert: mongo.NewUpdateOneModel().Upsert})
	if err != nil {
		return err
	}
	return nil
}

func UpdateMany(collection string, filter bson.D, update bson.D) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func ReplaceOne(collection string, filter interface{}, replacement interface{}) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	opts := options.Replace().SetUpsert(true)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.ReplaceOne(context.TODO(), filter, replacement, opts)
	if err != nil {
		return err
	}
	return nil

}

func DeleteOne(collection string, filter bson.D) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMany(collection string, filter bson.D) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	_, err = coll.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func FindOne(collection string, filter bson.D, doc interface{}) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	// filter := bson.D{{"name", "Bagels N Buns"}}
	err = coll.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return err
		}
		return err
	}
	return nil
}
func FindOneWorkflow(collection string, filter bson.D, doc internaltypes.Workflow) error {
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	// filter := bson.D{{"name", "Bagels N Buns"}}
	err = coll.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return err
		}
		return err
	}
	return nil
}

func FindMany(collection string, filter bson.D) ([]interface{}, error) {
	var docs []interface{}
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	// filter := bson.D{{"name", "Bagels N Buns"}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return docs, err
	}
	if err = cursor.All(context.TODO(), &docs); err != nil {
		return docs, err
	}

	// for _, doc := range docs {
	// 	err := cursor.Decode(&doc)
	// 	if err != nil {
	// 		return docs, err
	// 	}

	// }
	return docs, err
}

func FindManyResults(collection string, filter bson.D) ([]internaltypes.Result, error) {
	var docs []internaltypes.Result
	clientOptions := options.Client().ApplyURI(viper.GetString("mongodb_uri"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	logger.HandleErr(err)
	coll := client.Database("opsilon").Collection(collection)
	// filter := bson.D{{"name", "Bagels N Buns"}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return docs, err
	}
	if err = cursor.All(context.TODO(), &docs); err != nil {
		return docs, err
	}

	// for _, doc := range docs {
	// 	err := cursor.Decode(&doc)
	// 	if err != nil {
	// 		return docs, err
	// 	}

	// }
	return docs, err
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
