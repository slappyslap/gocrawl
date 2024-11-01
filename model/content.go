package model

import (
	"GoCrawl/internal/crypto"
	"GoCrawl/internal/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Header struct {
	Key   string
	Value string
}

type Content struct {
	Id          string    `bson:"_id,omitempty" json:"_id,omitempty"`
	Domain      string    `bson:"domain" json:"domain"`
	URL         string    `bson:"url" json:"url"`
	Title       string    `bson:"title" json:"title"`
	Desc        string    `bson:"desc" json:"desc"`
	Author      string    `bson:"author" json:"author"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
	HttpHeaders []Header  `bson:"http_headers" json:"http_headers"`
}

func GetContentCount() (int64, error) {
	opts := options.Count().SetHint("_id_")
	return content.CountDocuments(ctx, bson.D{}, opts)
}

func (m *Content) Insert() error {
	m.Id = crypto.Md5(m.URL)
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	_, err := content.InsertOne(ctx, m)

	if err != nil {
		log.Error("failed to insert content, err:%s", err.Error())
		return err
	}

	return nil
}

func InsertMultipleContents(contents []Content) error {

	log.Info("Inserting %d contents", len(contents))

	var docs []interface{}

	for _, c := range contents {
		c.Id = crypto.Md5(c.URL)
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()
		docs = append(docs, c)
	}

	_, err := content.InsertMany(ctx, docs)

	if err != nil {
		log.Error("failed to insert content, err:%s", err.Error())
		return err
	}

	return nil
}
