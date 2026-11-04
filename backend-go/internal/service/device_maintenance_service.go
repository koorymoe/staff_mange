package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type DeviceMaintenanceService struct {
	repo         *repository.DeviceMaintenanceRepository
	customerRepo *repository.CustomerRepository
	durations    *JobDurationEstimatorService
}

func NewDeviceMaintenanceService(
	repo *repository.DeviceMaintenanceRepository,
	customerRepo *repository.CustomerRepository,
	durations *JobDurationEstimatorService,
) *DeviceMaintenanceService {
	return &DeviceMaintenanceService{repo: repo, customerRepo: customerRepo, durations: durations}
}

func (s *DeviceMaintenanceService) Create(employeeID string, req model.CreateDeviceMaintenanceTicketRequest) (*model.DeviceMaintenanceTicket, error) {
	if req.DeviceTypeName == "" || req.Problem == "" {
		return nil, errors.New("نوع/اسم الجهاز والمشكلة مطلوبان")
	}
	customer, err := s.customerRepo.FindByCode(req.CustomerCode)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errors.New("لا يوجد زبون بهذا الكود")
	}
	return s.repo.Create(employeeID, customer.ID, req)
}

func (s *DeviceMaintenanceService) List() ([]model.DeviceMaintenanceTicket, error) {
	return s.repo.List()
}

func (s *DeviceMaintenanceService) Update(id string, req model.UpdateDeviceMaintenanceTicketRequest) (*model.DeviceMaintenanceTicket, error) {
	wasDelivered := false
	if req.MarkDelivered {
		if before, err := s.repo.GetByID(id); err == nil && before != nil {
			wasDelivered = before.DeliveredAt != nil
		}
	}

	ticket, err := s.repo.Update(id, req)
	if err != nil {
		return nil, err
	}

	// تسجيل عيّنة زمن تنفيذ حقيقية عند أول تسليم فعلي فقط (مو عند كل تعديل لاحق) —
	// itemCount=1 (جهاز واحد)، crewSize=1 (صيانة الأجهزة فردية، ما فيها مفهوم فريق حالياً).
	if req.MarkDelivered && !wasDelivered && ticket != nil && ticket.DeliveredAt != nil && ticket.ReceivedAt != nil && s.durations != nil {
		durationMinutes := int(ticket.DeliveredAt.Sub(*ticket.ReceivedAt).Minutes())
		if durationMinutes > 0 {
			ticketID := ticket.ID
			_ = s.durations.RecordSample(model.JobDurationSample{
				SystemName:                ticket.DeviceTypeName,
				JobType:                   model.JobTypeMaintenance,
				ItemCount:                 1,
				CrewSize:                  1,
				DurationMinutes:           durationMinutes,
				DeviceMaintenanceTicketID: &ticketID,
			})
			employeeName := ticket.DeviceTypeName
			if ticket.Employee != nil {
				employeeName = ticket.Employee.Name
			}
			_ = s.durations.CheckOverrunAndNotify(
				ticket.DeviceTypeName, model.JobTypeMaintenance, 1, 1, durationMinutes,
				employeeName, "صيانة "+ticket.DeviceTypeName, nil,
			)
		}
	}

	return ticket, nil
}
