package usecase

import "github.com/javaman/go-metrics/internal/domain"

type metricUsecase struct {
	metricRepo domain.MetricRepository
}

func New(m domain.MetricRepository) *metricUsecase {
	return &metricUsecase{metricRepo: m}
}

func (muc *metricUsecase) Save(m *domain.Metric) (*domain.Metric, error) {
	switch m.MType {
	case domain.Gauge:
		if err := muc.metricRepo.Save(m); err != nil {
			return nil, err
		}
		value := *m.Value
		return &domain.Metric{ID: m.ID, MType: m.MType, Value: &value}, nil
	case domain.Counter:
		var result *domain.Metric
		var err error
		if result, err = muc.metricRepo.Get(m); err != nil {
			var delta int64
			result = &domain.Metric{ID: m.ID, MType: m.MType, Delta: &delta}
		}
		*result.Delta += *m.Delta
		if err := muc.metricRepo.Save(result); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, domain.ErrorInvalidType
	}
}

func (muc *metricUsecase) SaveAll(ms []domain.Metric) error {
	for _, metric := range ms {
		if _, err := muc.Save(&metric); err != nil {
			return err
		}
	}
	return nil
}

func (muc *metricUsecase) Get(m *domain.Metric) (*domain.Metric, error) {
	return muc.metricRep.Get(m)
}

func (muc *metricUsecase) List() ([]*domain.Metric, error) {
	return muc.metricRep.List()
}

func (muc *metricUsecase) Ping() bool {
	return muc.metricRep.Ping()
}
