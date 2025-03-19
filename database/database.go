package database

import (
	"context"
	"fmt"
	"time"

	"github.com/natnael-wondwoesn/crud_golang/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var connectionString string = "mongodb+srv://natiwonde:wGr9c9dSnduTTBd@cluster0.aw06t.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"

type DB struct {
	client *mongo.Client
}

func Connect() (*DB, error) {
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected and pinged.")
	return &DB{
		client: client,
	}, nil
}

func (db *DB) GetJob(id string) *model.JobListing {
	jobColl := db.client.Database("joblistings").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("Invalid job ID format:", err)
		return nil
	}
	filter := bson.M{"_id": _id}
	var jobListing model.JobListing
	err = jobColl.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil {
		fmt.Println("Error finding job:", err)
		if err == mongo.ErrNoDocuments {
            return nil
        }
		return nil
	}
	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobColl := db.client.Database("joblistings").Collection("jobs")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var jobListings []*model.JobListing
	cursor, err := jobColl.Find(ctx, bson.D{})
	if err != nil {
		fmt.Println("Error finding jobs:", err)
		return nil
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var jobListing model.JobListing
		if err := cursor.Decode(&jobListing); err != nil {
			fmt.Println("Error decoding job:", err)
			return nil
		}
		jobListings = append(jobListings, &jobListing)
	}

	if err := cursor.Err(); err != nil {
		fmt.Println("Cursor error:", err)
		return nil
	}

	return jobListings

}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobColl := db.client.Database("joblistings").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	newJob := bson.M{
        "title":       jobInfo.Title,
        "description": jobInfo.Description,
        "company":     jobInfo.Company,
		"url" : jobInfo.URL,
    }

	insert, err := jobColl.InsertOne(ctx, newJob)
	if err != nil {
		fmt.Println("Error creating job:", err)
		return nil
	}

	returnJobList := model.JobListing{ID: insert.InsertedID.(primitive.ObjectID).Hex(), Title: jobInfo.Title, Description: jobInfo.Description, Company: jobInfo.Company, URL:jobInfo.URL}
	return &returnJobList
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobColl := db.client.Database("joblistings").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	updateJobInfo := bson.M{}

	if jobInfo.Title != nil {
		updateJobInfo["title"] = *jobInfo.Title
	}
	if jobInfo.Description != nil {
		updateJobInfo["description"] = *jobInfo.Description
	}
	if jobInfo.Company != nil {
		updateJobInfo["company"] = *jobInfo.Company
	}
	if jobInfo.URL != nil {
		updateJobInfo["url"] = *jobInfo.URL
	}

	_id, err := primitive.ObjectIDFromHex(jobId)

    if err != nil {
		fmt.Println("Invalid ID format:", err)
        return nil
    }

	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJobInfo}

	results := jobColl.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))
	var jobListing model.JobListing
	if err := results.Decode(&jobListing); err != nil {
		fmt.Println("Error updating job:", err)
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return nil
	}

	return &jobListing
}

func (db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse {
	jobColl := db.client.Database("joblistings").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		fmt.Println("Invalid job ID format:", err)
		return &model.DeleteJobResponse{Success: false}
	}

	filter := bson.M{"_id": _id}

	result,err := jobColl.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Println("Error deleting Job:",err)
		return &model.DeleteJobResponse{Success: false}
	}
	if result.DeletedCount == 0 {
		fmt.Println("Job listing not found")
		return &model.DeleteJobResponse{Success: false}
	}

	return &model.DeleteJobResponse{Success: true}

}
