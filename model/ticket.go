package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Ticket status constants
const (
	TicketStatusPending    = "pending"
	TicketStatusProcessing = "processing"
	TicketStatusResolved   = "resolved"
	TicketStatusClosed     = "closed"
)

// Ticket priority constants
const (
	TicketPriorityLow    = "low"
	TicketPriorityMedium = "medium"
	TicketPriorityHigh   = "high"
	TicketPriorityUrgent = "urgent"
)

// Ticket category constants
const (
	TicketCategoryTechnical = "technical"
	TicketCategoryAccount   = "account"
	TicketCategoryFeature   = "feature"
	TicketCategoryOther     = "other"
)

type Ticket struct {
	Id          int            `json:"id" gorm:"primaryKey"`
	UserId      int            `json:"user_id" gorm:"index;not null"`
	Username    string         `json:"username" gorm:"-"` // Virtual field for frontend display
	Title       string         `json:"title" gorm:"type:varchar(200);not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Status      string         `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	Priority    string         `json:"priority" gorm:"type:varchar(20);default:'medium'"`
	Category    string         `json:"category" gorm:"type:varchar(50);default:'other'"`
	Attachments string         `json:"attachments" gorm:"type:text"` // JSON array of attachment URLs
	AdminReply  string         `json:"admin_reply" gorm:"type:text"`
	RepliedBy   *int           `json:"replied_by"`
	RepliedAt   *time.Time     `json:"replied_at"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ticket *Ticket) Insert() error {
	return DB.Create(ticket).Error
}

func (ticket *Ticket) Update() error {
	return DB.Save(ticket).Error
}

func (ticket *Ticket) Delete() error {
	return DB.Delete(ticket).Error
}

func GetTicketById(id int) (*Ticket, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	ticket := &Ticket{}
	err := DB.First(ticket, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	
	// Load username
	user, err := GetUserById(ticket.UserId, false)
	if err == nil {
		ticket.Username = user.Username
	}
	
	return ticket, nil
}

func GetTicketsByUserId(userId int, pageNum int, pageSize int) ([]*Ticket, int64, error) {
	var tickets []*Ticket
	var total int64
	
	offset := (pageNum - 1) * pageSize
	
	err := DB.Model(&Ticket{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = DB.Where("user_id = ?", userId).
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&tickets).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	// Load usernames
	for i := range tickets {
		user, err := GetUserById(tickets[i].UserId, false)
		if err == nil {
			tickets[i].Username = user.Username
		}
	}
	
	return tickets, total, nil
}

func GetAllTickets(pageNum int, pageSize int, status string) ([]*Ticket, int64, error) {
	var tickets []*Ticket
	var total int64
	
	offset := (pageNum - 1) * pageSize
	
	query := DB.Model(&Ticket{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	err = query.Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&tickets).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	// Load usernames
	for i := range tickets {
		user, err := GetUserById(tickets[i].UserId, false)
		if err == nil {
			tickets[i].Username = user.Username
		}
	}
	
	return tickets, total, nil
}

func UpdateTicketStatus(id int, status string) error {
	return DB.Model(&Ticket{}).Where("id = ?", id).Update("status", status).Error
}

func UpdateTicketReply(id int, reply string, adminId int) error {
	now := time.Now()
	return DB.Model(&Ticket{}).Where("id = ?", id).Updates(map[string]interface{}{
		"admin_reply": reply,
		"replied_by":  adminId,
		"replied_at":  now,
	}).Error
}
