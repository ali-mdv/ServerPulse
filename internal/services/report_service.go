package services

import (
	"server-monitoring/internal/repository"
	database "server-monitoring/pkg/mongo"
)

type ReportService interface {
	ReportsList() any
}

type reportService struct {
	repo repository.ReportRepository
}

func NewReportService(dbName string) ReportService {
	db := database.GetDatabase(dbName)
	reportRepo := repository.NewReportRepository(db)
	return &reportService{repo: reportRepo}
}

func (s *reportService) ReportsList() any {
	return s.repo.ReportsList()
}
