package db_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	// "reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transaction[DocType interface{}] interface {
	Commit() error
	Rollback() error
	CreateDocument(ctx context.Context, id string, document *DocType) error
}

type mongoTransaction[DocType interface{}] struct {
	session    mongo.Session
	DbName     string
	Collection string
	Timeout    time.Duration
}

func (this *mongoTransaction[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()

	db := this.session.Client().Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil: // no error means there is conflicting document
		return ErrConflict
	case mongo.ErrNoDocuments:
		// do nothing, this is expected
	default: // other errors - return them
		return result.Err()
	}

	_, err := collection.InsertOne(ctx, document)
	return err
}

func (t *mongoTransaction[DocType]) Commit() error {
	err := t.session.CommitTransaction(context.Background())
	t.session.EndSession(context.Background())
	return err
}

func (t *mongoTransaction[DocType]) Rollback() error {
	err := t.session.AbortTransaction(context.Background())
	t.session.EndSession(context.Background())
	return err
}

type DbService[DocType interface{}] interface {
	CreateDocument(ctx context.Context, id string, document *DocType) error
	FindDocument(ctx context.Context, id string) (*DocType, error)
	FindDocuments(ctx context.Context, filter interface{}) ([]*DocType, error)
	UpdateDocument(ctx context.Context, id string, document *DocType) error
	DeleteDocument(ctx context.Context, id string) error
	BeginTransaction(ctx context.Context) (Transaction[DocType], error)
	Disconnect(ctx context.Context) error
}

var ErrNotFound = fmt.Errorf("document not found")
var ErrConflict = fmt.Errorf("conflict: document already exists")

type MongoServiceConfig struct {
	ServerHost string
	ServerPort int
	UserName   string
	Password   string
	DbName     string
	Collection string
	Timeout    time.Duration
}

type mongoSvc[DocType interface{}] struct {
	MongoServiceConfig
	client     atomic.Pointer[mongo.Client]
	clientLock sync.Mutex
}

func NewMongoService[DocType interface{}](config MongoServiceConfig) DbService[DocType] {
	enviro := func(name string, defaultValue string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		return defaultValue
	}

	svc := &mongoSvc[DocType]{}
	svc.MongoServiceConfig = config

	if svc.ServerHost == "" {
		svc.ServerHost = enviro("API_MONGODB_HOST", "localhost")
	}

	if svc.ServerPort == 0 {
		port := enviro("API_MONGODB_PORT", "27017")
		if port, err := strconv.Atoi(port); err == nil {
			svc.ServerPort = port
		} else {
			log.Printf("Invalid port value: %v", port)
			svc.ServerPort = 27017
		}
	}

	if svc.UserName == "" {
		svc.UserName = enviro("API_MONGODB_USERNAME", "")
	}

	if svc.Password == "" {
		svc.Password = enviro("API_MONGODB_PASSWORD", "")
	}

	if svc.DbName == "" {
		svc.DbName = enviro("API_MONGODB_DATABASE", "ss-sprava-krvi")
	}

	if svc.Collection == "" {
		svc.Collection = enviro("API_MONGODB_COLLECTION", "donor")
	}

	if svc.Timeout == 0 {
		seconds := enviro("API_MONGODB_TIMEOUT_SECONDS", "10")
		if seconds, err := strconv.Atoi(seconds); err == nil {
			svc.Timeout = time.Duration(seconds) * time.Second
		} else {
			log.Printf("Invalid timeout value: %v", seconds)
			svc.Timeout = 10 * time.Second
		}
	}

	log.Printf(
		"MongoDB config: //%v@%v:%v/%v/%v",
		svc.UserName,
		svc.ServerHost,
		svc.ServerPort,
		svc.DbName,
		svc.Collection,
	)
	return svc
}

func (this *mongoSvc[DocType]) connect(ctx context.Context) (*mongo.Client, error) {
	// optimistic check
	client := this.client.Load()
	if client != nil {
		return client, nil
	}

	this.clientLock.Lock()
	defer this.clientLock.Unlock()
	// pesimistic check
	client = this.client.Load()
	if client != nil {
		return client, nil
	}

	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()

	var uri = fmt.Sprintf("mongodb://%v:%v", this.ServerHost, this.ServerPort)
	log.Printf("Using URI: " + uri)

	if len(this.UserName) != 0 {
		uri = fmt.Sprintf("mongodb://%v:%v@%v:%v", this.UserName, this.Password, this.ServerHost, this.ServerPort)
	}

	if client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(10*time.Second)); err != nil {
		return nil, err
	} else {
		this.client.Store(client)
		return client, nil
	}
}

func (this *mongoSvc[DocType]) Disconnect(ctx context.Context) error {
	client := this.client.Load()

	if client != nil {
		this.clientLock.Lock()
		defer this.clientLock.Unlock()

		client = this.client.Load()
		defer this.client.Store(nil)
		if client != nil {
			if err := client.Disconnect(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func GetDbService[DocType interface{}](ctx context.Context, ctxKey string) (DbService[DocType], error) {
	value := ctx.Value(ctxKey)
	if value == nil {
		return nil, errors.New("db_service_donors not found")
	}

	db, ok := value.(DbService[DocType])
	if !ok {
		return nil, errors.New("cannot cast db context to db_service.DbService")
	}

	return db, nil
}
func (this *mongoSvc[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil: // no error means there is conflicting document
		return ErrConflict
	case mongo.ErrNoDocuments:
		// do nothing, this is expected
	default: // other errors - return them
		return result.Err()
	}

	_, err = collection.InsertOne(ctx, document)
	return err
}

func (this *mongoSvc[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return nil, err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default: // other errors - return them
		return nil, result.Err()
	}
	var document *DocType
	if err := result.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}

func (this *mongoSvc[DocType]) FindDocuments(ctx context.Context, filter interface{}) ([]*DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return nil, err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)

	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return nil, errors.New("could not process filters")
	}
	var bsonFilter bson.D
	err = bson.UnmarshalExtJSON(filterBytes, true, &bsonFilter)
	if err != nil {
		return nil, errors.New("could not process filters")
	}
	// log.Printf("bson filters: %v", bsonFilter)
	result, err := collection.Find(ctx, bsonFilter)
	switch err {
	case nil:
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default: // other errors - return them
		return nil, err
	}
	var documents []*DocType
	for result.Next(ctx) {
		var document *DocType
		if err := result.Decode(&document); err != nil {
			return nil, errors.New("some of the documents could not be read")
		}
		documents = append(documents, document)
	}
	if len(documents) == 0 {
		return []*DocType{}, nil

	}
	return documents, nil
}

func (this *mongoSvc[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default: // other errors - return them
		return result.Err()
	}

	// Preserve created_at
	// var rawDocument bson.Raw
	// if err := result.Decode(&rawDocument); err != nil {
	// 	return err
	// }
	// log.Printf("raw document: %v", rawDocument)
	// if CreatedAt, found := rawDocument.Lookup("createdat").StringValueOK(); found {
	// 	field := reflect.ValueOf(document).Elem().FieldByName("createdAt")
	// 	if field.IsValid() && field.CanSet() {
	// 		switch field.Kind() {
	// 		case reflect.String:
	// 			field.SetString(CreatedAt)
	// 		default:
	// 			panic("failed to preserve created_at") // TODO: logging
	// 		}
	// 	}
	// }

	// field := reflect.ValueOf(document).Elem().FieldByName("UpdatedAt")
	// if field.IsValid() && field.CanSet() {
	// 	switch field.Kind() {
	// 	case reflect.String:
	// 		field.SetString(time.Now().Format("Sat Jan 01 2022 00:00:00 GMT+0000 (Coordinated Universal Time)"))
	// 	default:
	// 		log.Printf("raw document: %v", field.Kind)
	// 		panic("failed to set updated_at") // TODO: logging
	// 	}
	// }

	_, err = collection.ReplaceOne(ctx, bson.D{{Key: "id", Value: id}}, document)
	return err
}

func (this *mongoSvc[DocType]) DeleteDocument(ctx context.Context, id string) error {
	ctx, contextCancel := context.WithTimeout(ctx, this.Timeout)
	defer contextCancel()
	client, err := this.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(this.DbName)
	collection := db.Collection(this.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default: // other errors - return them
		return result.Err()
	}
	_, err = collection.DeleteOne(ctx, bson.D{{Key: "id", Value: id}})
	return err
}

func (this *mongoSvc[DocType]) BeginTransaction(ctx context.Context) (Transaction[DocType], error) {
	client, err := this.connect(ctx)
	if err != nil {
		return nil, err
	}

	session, err := client.StartSession()
	if err != nil {
		return nil, err
	}

	return &mongoTransaction[DocType]{
		session:    session,
		Timeout:    this.Timeout,
		DbName:     this.DbName,
		Collection: this.Collection,
	}, nil
}
