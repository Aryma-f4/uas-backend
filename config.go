package main

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
)

func InitPostgres() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/achievement_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func InitMongoDB() (*mongo.Client, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	return client, nil
}
