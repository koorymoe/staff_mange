package model

import "time"

type Booking struct {
	ID                     string     `db:"id" json:"id"`
	Code                   string     `db:"code" json:"code"`
	SequenceNumber         *int       `db:"sequenceNumber" json:"sequenceNumber"`
	ScheduledAt            *time.Time `db:"scheduledAt" json:"scheduledAt"`
	PendingScheduledAt     *time.Time `db:"pendingScheduledAt" json:"pendingScheduledAt"`
	CustomerID             string     `db:"customerId" json:"-"`
	ServiceID              *string    `db:"serviceId" json:"-"`
	TransferEmployeeID     *string    `db:"transferEmployeeId" json:"-"`
	ProjectSupervisorID    *string    `db:"projectSupervisorId" json:"-"`
	ExpenseResponsibleID   *string    `db:"expenseResponsibleId" json:"expenseResponsibleId"`
	ConfirmedByEmployeeID  *string    `db:"confirmedByEmployeeId" json:"-"`
	Notes                  *string    `db:"notes" json:"notes"`
	VehicleType            *string    `db:"vehicleType" json:"vehicleType"`
	Priority               string     `db:"priority" json:"priority"`
	Status                 string     `db:"status" json:"status"`
	TransferToProjects     bool       `db:"transferToProjects" json:"transferToProjects"`
	ConfirmedByName        *string    `db:"confirmedByName" json:"confirmedByName"`
	AdminNotes             *string    `db:"adminNotes" json:"adminNotes"`
	AssignedVehicle        *string    `db:"assignedVehicle" json:"assignedVehicle"`
	QuotedPrice            *float64   `db:"quotedPrice" json:"quotedPrice"`
	Address                *string    `db:"address" json:"address"`
	MapLocation            *string    `db:"mapLocation" json:"mapLocation"`
	MapLatitude            *float64   `db:"mapLatitude" json:"mapLatitude"`
	MapLongitude           *float64   `db:"mapLongitude" json:"mapLongitude"`
	CompletedAt            *time.Time `db:"completedAt" json:"completedAt"`
	CompletionNotes        *string    `db:"completionNotes" json:"completionNotes"`
	AmountCollected        *float64   `db:"amountCollected" json:"amountCollected"`
	AdvancePaid            *float64   `db:"advancePaid" json:"advancePaid"`
	AmountVerified         bool       `db:"amountVerified" json:"amountVerified"`
	EquipmentStatus        string     `db:"equipmentStatus" json:"equipmentStatus"`
	Shift                  *string    `db:"shift" json:"shift"`
	DeviceCount            *int       `db:"deviceCount" json:"deviceCount"`
	InspectionSupervisorID *string    `db:"inspectionSupervisorId" json:"inspectionSupervisorId"`
	ProjectCar             *string    `db:"projectCar" json:"projectCar"`
	CrewNotes              *string    `db:"crewNotes" json:"crewNotes"`
	BookingType            string     `db:"bookingType" json:"bookingType"`
	Urgency                *string    `db:"urgency" json:"urgency"`
	MaintenanceType        *string    `db:"maintenanceType" json:"maintenanceType"`
	RemembersExecutionCrew bool       `db:"remembersExecutionCrew" json:"remembersExecutionCrew"`
	SystemCount            *int       `db:"systemCount" json:"systemCount"`
	SystemType             *string    `db:"systemType" json:"systemType"`
	ProjectSpeed           *string    `db:"projectSpeed" json:"projectSpeed"`
	WorkType               *string    `db:"workType" json:"workType"`
	AddressDescription     *string    `db:"addressDescription" json:"addressDescription"`
	CreatedAt              time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time  `db:"updatedAt" json:"updatedAt"`

	Customer            *Customer           `db:"-" json:"customer"`
	Service             *Service            `db:"-" json:"service"`
	TransferEmployee    *Employee           `db:"-" json:"transferEmployee"`
	ProjectSupervisor   *Employee           `db:"-" json:"projectSupervisor"`
	ConfirmedByEmployee *Employee           `db:"-" json:"confirmedByEmployee"`
	ExpenseResponsible  *Employee           `db:"-" json:"expenseResponsible"`
	Assignments         []BookingAssignment `db:"-" json:"assignments"`
	CartItems           []CartItem          `db:"-" json:"cartItems"`
	ScheduleLogs        []ScheduleChangeLog `db:"-" json:"scheduleLogs"`
}

type BookingAssignment struct {
	ID         string    `db:"id" json:"id"`
	BookingID  string    `db:"bookingId" json:"-"`
	EmployeeID string    `db:"employeeId" json:"-"`
	Role       string    `db:"role" json:"role"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	Employee   Employee  `db:"-" json:"employee"`
}

type CartItem struct {
	ID          string    `db:"id" json:"id"`
	BookingID   string    `db:"bookingId" json:"bookingId"`
	ProductName string    `db:"productName" json:"productName"`
	Quantity    float64   `db:"quantity" json:"quantity"`
	UnitPrice   float64   `db:"unitPrice" json:"unitPrice"`
	TotalPrice  float64   `db:"totalPrice" json:"totalPrice"`
	Notes       *string   `db:"notes" json:"notes"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt" json:"updatedAt"`
}

type ScheduleChangeLog struct {
	ID          string     `db:"id" json:"id"`
	BookingID   string     `db:"bookingId" json:"-"`
	ChangedByID string     `db:"changedById" json:"changedById"`
	OldTime     *time.Time `db:"oldTime" json:"oldTime"`
	NewTime     time.Time  `db:"newTime" json:"newTime"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	ChangedBy   *Employee  `db:"-" json:"changedBy"`
}

type CreateBookingRequest struct {
	CustomerID         string   `json:"customerId"`
	ServiceID          *string  `json:"serviceId"`
	Notes              *string  `json:"notes"`
	VehicleType        *string  `json:"vehicleType"`
	Priority           *string  `json:"priority"`
	TransferEmployeeID *string  `json:"transferEmployeeId"`
	Address            *string  `json:"address"`
	MapLatitude        *float64 `json:"mapLatitude"`
	MapLongitude       *float64 `json:"mapLongitude"`
}

type ConfirmBookingRequest struct {
	ConfirmedByName       *string  `json:"confirmedByName"`
	ConfirmedByEmployeeID *string  `json:"confirmedByEmployeeId"`
	AdminNotes            *string  `json:"adminNotes"`
	TransferToProjects    bool     `json:"transferToProjects"`
	QuotedPrice           *float64 `json:"quotedPrice"`
	Address               *string  `json:"address"`
	ScheduledAt           *string  `json:"scheduledAt"`
}

type AssignBookingRequest struct {
	EmployeeID      string  `json:"employeeId"`
	Role            string  `json:"role"`
	AssignedVehicle *string `json:"assignedVehicle"`
}

type CreateCartItemRequest struct {
	ProductName string   `json:"productName"`
	Quantity    *float64 `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	Notes       *string  `json:"notes"`
}

type UpdateCartItemRequest struct {
	ProductName *string  `json:"productName"`
	Quantity    *float64 `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	Notes       *string  `json:"notes"`
}

type UpdateBookingDetailsRequest struct {
	QuotedPrice          *float64 `json:"quotedPrice"`
	Address              *string  `json:"address"`
	AssignedVehicle      *string  `json:"assignedVehicle"`
	MapLocation          *string  `json:"mapLocation"`
	MapLatitude          *float64 `json:"mapLatitude"`
	MapLongitude         *float64 `json:"mapLongitude"`
	ExpenseResponsibleID *string  `json:"expenseResponsibleId"`
}

type CompleteBookingRequest struct {
	CompletionNotes *string  `json:"completionNotes"`
	AmountCollected *float64 `json:"amountCollected"`
	AdvancePaid     *float64 `json:"advancePaid"`
}
