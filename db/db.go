package db

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
	"log"
	"context"
	"errors"
	"fmt"
	//"time"
	//"bytes"
	//"encoding/json"
	"time"
)

const (
	DB_HOST = "mongo"
	DB_USER = "root"
	DB_PASS = "mongotest"
	DB_NAME = "analyzer"
	DB_COLLECTION = "packets"
	DB_PORT = 27017
)

type TransportType uint8

type PacketWrapper struct {
	Device 		string			`json:"device" bson:"Device"`
	Type		float64			`json:"type" bson:"Type"`
	SrcIP		string/*net.IP*/`json:"src_ip" bson:"SrcIP"`
	DstIP		string/*net.IP*/`json:"dst_ip" bson:"DstIP"`
	SrcPort		string			`json:"src_port" bson:"SrcPort"`
	DstPort		string			`json:"dst_port" bson:"DstPort"`
	Payload 	string			`json:"payload" bson:"Payload"`
	CapturedAt	int64			`json:"captured_at" bson:"CapturedAt"`
}

func Connect() (*mongo.Client, error)  {

	connString := fmt.Sprintf("mongodb://%s:%s@%s:%v", DB_USER, DB_PASS, DB_HOST, DB_PORT)

	client, error := mongo.Connect(context.Background(), connString, nil)//mongo.NewClient(connString)

	//if error != nil {
	//	log.Fatal(error)
	//}

	//error = client.Connect(context.TODO())

	if error != nil {
		return nil, errors.New(fmt.Sprintf("error connecting to DB using this connection string: %s", connString))
	}

	return client, nil
}

func Save(doc map[string]interface{}) {
	log.Printf("%v", doc)

	client, err := Connect()

	if err != nil {
		log.Fatal(err)
	}

	t, err := time.Parse(time.RFC3339, doc["CapturedAt"].(string))

	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(DB_NAME).Collection(DB_COLLECTION)

	res, err := collection.InsertOne(context.Background(),
		bson.NewDocument(
			bson.EC.Double("Type" , doc["Type"].(float64)),
			bson.EC.String("SrcIP", doc["SrcIP"].(string)),
			bson.EC.String("DstIP", doc["DstIP"].(string)),
			bson.EC.String("SrcPort", doc["SrcPort"].(string)),
			bson.EC.String("DstPort", doc["DstPort"].(string)),
			bson.EC.String("Device", doc["Device"].(string)),
			bson.EC.Int64("CapturedAt", t.UnixNano()),
			bson.EC.String("Payload", doc["Payload"].(string)),
			),
	)

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID : %v", res.InsertedID)

	client.Disconnect(context.Background())
}

func FindAll() ([]PacketWrapper, error) {

	client, err := Connect()

	if err != nil {
		log.Fatal(err)
	}

	collection  := client.Database(DB_NAME).Collection(DB_COLLECTION)

	//cur, err := collection.Find(context.Background(), nil)

	pipeline := bson.NewArray(
		bson.VC.DocumentFromElements(
			bson.EC.SubDocumentFromElements(
				"$sort",
				bson.EC.Int32("CreatedAt", 1),
			),
		),
	)
	cur, err := collection.Aggregate(context.Background(), pipeline)

	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())


	var l []PacketWrapper

	for cur.Next(context.Background()) {
		elem := PacketWrapper{}
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		//a := bson.NewDocument()
		//err = cur.Decode(&a)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//log.Printf("%v", a)
		// do something with elem....

		//s, err := bson.Unmarshal(elem)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//log.Printf("%v", s)
		l = append(l, elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	//log.Printf("Packets: %v", l)
	client.Disconnect(context.Background())

	return l, nil

}