package backend

import (
	"context"
	"fmt"

	"time"

	cds "cloud.google.com/go/datastore"
	"github.com/pkg/errors"
	metadatabox "github.com/sinmetal/gcpbox/metadata"
	"github.com/vvakame/sdlog/aelog"
	"go.mercari.io/datastore"
	"go.mercari.io/datastore/clouddatastore"
)

var _ datastore.PropertyLoadSaver = &SpannerAccount{}
var _ datastore.KeyLoader = &SpannerAccount{}

// SpannerAccount is Spanner利用者Account Datastore Entity Model
// +qbg
type SpannerAccount struct {
	Key             datastore.Key `datastore:"-" json:"-"`
	KeyStr          string        `datastore:"-" json:"key"`
	GCPUGSlackID    string        `json:"gcpugSlackId"`
	ServiceAccounts []string      `json:"serviceAccounts"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	SchemaVersion   int           `json:"-"`
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
	pID, err := metadatabox.ProjectID()
	if err != nil {
		aelog.Errorf(ctx, "failed get project id: %+v", err)
		return nil, err
	}
	cdsClient, err := cds.NewClient(ctx, pID)
	if err != nil {
		aelog.Errorf(ctx, "failed get project id: %+v", err)
		return nil, err
	}
	ds, err := clouddatastore.FromClient(ctx, cdsClient)
	if err != nil {
		aelog.Errorf(ctx, "failed Datastore New Client: %+v", err)
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
func (store *SpannerAccountStore) Upsert(ctx context.Context, key datastore.Key, e *SpannerAccount) (*SpannerAccount, error) {
	ds := store.DatastoreClient

	var se SpannerAccount
	_, err := ds.RunInTransaction(ctx, func(tx datastore.Transaction) error {
		if err := ds.Get(ctx, key, &se); err != nil {
			if err != datastore.ErrNoSuchEntity {
				return errors.Wrap(err, fmt.Sprintf("failed get SpannerAccount from Datastore. key=%v", key))
			}
		}

		sam := make(map[string]string)
		for _, sa := range se.ServiceAccounts {
			sam[sa] = sa
		}

		se.GCPUGSlackID = e.GCPUGSlackID
		for _, sa := range e.ServiceAccounts {
			aelog.Infof(ctx, "%s is request SA", sa)
			if _, ok := sam[sa]; !ok {
				aelog.Infof(ctx, "%s is add", sa)
				se.ServiceAccounts = append(se.ServiceAccounts, sa)
			}
		}

		_, err := ds.Put(ctx, key, &se)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed put SpannerAccount to Datastore. key=%v", key))
		}
		se.Key = key
		se.KeyStr = key.Encode()

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed commit SpannerAccount to Datastore. key=%v", key))
	}

	return &se, nil
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
