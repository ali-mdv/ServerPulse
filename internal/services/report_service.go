package services

import "go.mongodb.org/mongo-driver/v2/mongo"

type ReportService interface {
	ReportsList() any
}

type reportService struct {
	db *mongo.Collection
}

func NewReportService(db *mongo.Collection) ReportService {
	return &reportService{db: db}
}

func (s *reportService) ReportsList() any {
	return []any{}
}
