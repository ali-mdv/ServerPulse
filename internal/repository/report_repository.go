package repository

import "go.mongodb.org/mongo-driver/v2/mongo"

type ReportRepository interface {
	ReportsList() any
}

type reportRepository struct {
	Collection *mongo.Collection
}

func NewReportRepository(db *mongo.Database) ReportRepository {
	return &reportRepository{
		Collection: db.Collection("report"),
	}
}

func (r *reportRepository) ReportsList() any {
	return []any{}
}
