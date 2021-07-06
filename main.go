package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jessevdk/go-flags"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type Options struct {
}

type Interface struct {
	redis *redis.Client
	mongo *mongo.Client
	opts  Options
	ctx   context.Context
}

func main() {
	opts := parseOpts()
	client := NewDB(opts)
	client.HabrCommentsGrabber()

}

func (i *Interface) HabrCommentsGrabber() {
	res, err := http.Get("https://habr.com/ru/post/188010/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var name string
	doc.Find(".comment").Each(func(p int, comments *goquery.Selection) {
		comments.Find(".comment__head   ").Each(func(k int, head *goquery.Selection) {
			head.Find(".user-info").Each(func(f int, userInfo *goquery.Selection) {
				name = userInfo.Find(".user-info__nickname").Text()
			})
		})
		message := comments.Find(".comment__message  ").Text()
		if i.redis.Get(GetMD5Hash(name+message)).Val() != "1" {
			collection := i.mongo.Database("parser").Collection("parser")
			_, err = collection.InsertOne(i.ctx, bson.D{{"name", name}, {"message", message}})
			if err != nil {
				log.Println(err.Error())
			}
			i.redis.Set(GetMD5Hash(name+message), true, 0)
		} else {
			fmt.Println("Not new.")
		}

	})
	fmt.Println("Done.")
}

func parseOpts() Options {
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}
	return opts
}

func NewDB(opts Options) *Interface {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "mypass",
		DB:       0,
	})

	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatalf(err.Error())
	}

	return &Interface{
		redis: redisClient,
		opts:  opts,
		mongo: mongoClient,
		ctx:   ctx,
	}
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
