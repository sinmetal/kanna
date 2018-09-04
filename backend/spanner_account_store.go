package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mercari.io/datastore"
	"google.golang.org/appengine/log"
)

var _ datastore.PropertyLoadSaver = &SpannerAccount{}
var _ datastore.KeyLoader = &SpannerAccount{}

// SpannerAccount is Spanner利用者Account Datastore Entity Model
// +qbg
type SpannerAccount struct {
	Key             datastore.Key `datastore:"-" json:"-"`
	KeyStr          string        `datastore:"-" json:"key"`
	GCPUGSlackID string `json:"gcpugSlackId"`
	ServiceAccounts []string `json:"serviceAccounts"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	SchemaVersion   int `json:"-"`
}

// LoadKey is Entity Load時にKeyを設定する
func (e *SpannerAccount) LoadKey(ctx context.Context, k datastore.Key) error {
	e.Key = k

	return nil
}

// Load is Entity Load時に呼ばれる
func (e *SpannerAccount) Load(ctx context.Context, ps []datastore.Property) error {
	err := datastore.LoadStruct(ctx, e, ps)
	if err != nil {
		return err
	}

	return nil
}

// Save is Entity Save時に呼ばれる
func (e *SpannerAccount) Save(ctx context.Context) ([]datastore.Property, error) {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	e.UpdatedAt = time.Now()
	e.SchemaVersion = 1

	return datastore.SaveStruct(ctx, e)
}

// SpannerAccountStore is SpannerAccountのDatastoreの操作を司る
type SpannerAccountStore struct {
	DatastoreClient datastore.Client
}

// NewSpannerAccountStore is SpannerAccountStoreを作成
func NewSpannerAccountStore(ctx context.Context) (*SpannerAccountStore, error) {
	ds, err := FromContext(ctx)
	if err != nil {
		log.Errorf(ctx, "failed Datastore New Client: %+v", err)
		return nil, err
	}
	return &SpannerAccountStore{ds}, nil
}

// Kind is SpannerAccountのKindを返す
func (store *SpannerAccountStore) Kind() string {
	return "SpannerAccount"
}

// NameKey is SpannerAccountの指定したNameを利用したKeyを生成する
func (store *SpannerAccountStore) NameKey(ctx context.Context, name string) datastore.Key {
	return store.DatastoreClient.NameKey(store.Kind(), name, nil)
}

// Create is SpannerAccountをDatastoreにputする
func (store *SpannerAccountStore) Put(ctx context.Context, key datastore.Key, e *SpannerAccount) (*SpannerAccount, error) {
	ds := store.DatastoreClient

	_, err := ds.Put(ctx, key, e)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed put SpannerAccount to Datastore. key=%v", key))
	}
	e.Key = key
	e.KeyStr = key.Encode()
	return e, nil
}

// Get is SpannerAccountをDatastoreからgetする
func (store *SpannerAccountStore) Get(ctx context.Context, key datastore.Key) (*SpannerAccount, error) {
	ds := store.DatastoreClient

	var e SpannerAccount
	err := ds.Get(ctx, key, &e)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed get SpannerAccount from Datastore. key=%v", key))
	}
	e.KeyStr = key.Encode()

	return &e, nil
}
