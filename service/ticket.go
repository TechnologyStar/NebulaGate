package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	DefaultMaxTicketsPerDay = 5
)

// CheckTicketRateLimit checks if user has exceeded daily ticket creation limit
func CheckTicketRateLimit(userId int) (bool, error) {
	if !common.RedisEnabled {
		// If Redis is not enabled, allow ticket creation
		return true, nil
	}
	
	// Get max tickets per day from option or use default
	maxTicketsStr := common.OptionMapRWMutex.RLock()
	maxTicketsPerDay := common.OptionMap["MaxTicketsPerUserPerDay"]
	common.OptionMapRWMutex.RUnlock()
	
	maxTickets := DefaultMaxTicketsPerDay
	if maxTicketsPerDay != "" {
		if val, err := strconv.Atoi(maxTicketsPerDay); err == nil && val > 0 {
			maxTickets = val
		}
	}
	
	// Generate Redis key with current date
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("ticket:daily:%d:%s", userId, today)
	
	ctx := context.Background()
	
	// Increment counter
	count, err := common.RDB.Incr(ctx, key).Result()
	if err != nil {
		common.SysLog(fmt.Sprintf("Redis incr failed for ticket rate limit: %v", err))
		// If Redis fails, allow ticket creation to avoid blocking users
		return true, nil
	}
	
	// Set expiration on first increment
	if count == 1 {
		common.RDB.Expire(ctx, key, 24*time.Hour)
	}
	
	// Check if limit exceeded
	if count > int64(maxTickets) {
		return false, fmt.Errorf("已达到每日工单创建上限（%d个），请明天再试", maxTickets)
	}
	
	return true, nil
}

// CreateTicket creates a new ticket with validation and rate limiting
func CreateTicket(userId int, title, content, priority, category string) (*model.Ticket, error) {
	// Validate inputs
	if err := common.ValidateTicketTitle(title); err != nil {
		return nil, err
	}
	
	if err := common.ValidateTicketContent(content); err != nil {
		return nil, err
	}
	
	if !common.ValidateTicketPriority(priority) {
		return nil, errors.New("无效的优先级")
	}
	
	if !common.ValidateTicketCategory(category) {
		return nil, errors.New("无效的分类")
	}
	
	// Check rate limit
	allowed, err := CheckTicketRateLimit(userId)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errors.New("已达到每日工单创建上限")
	}
	
	// Sanitize inputs
	title = common.SanitizeInput(title)
	content = common.SanitizeInput(content)
	
	// Create ticket
	ticket := &model.Ticket{
		UserId:   userId,
		Title:    title,
		Content:  content,
		Status:   model.TicketStatusPending,
		Priority: priority,
		Category: category,
	}
	
	err = ticket.Insert()
	if err != nil {
		return nil, err
	}
	
	return ticket, nil
}

// GetUserTickets retrieves tickets for a specific user
func GetUserTickets(userId int, pageNum int, pageSize int) ([]*model.Ticket, int64, error) {
	return model.GetTicketsByUserId(userId, pageNum, pageSize)
}

// GetAllTicketsForAdmin retrieves all tickets (admin only)
func GetAllTicketsForAdmin(pageNum int, pageSize int, status string) ([]*model.Ticket, int64, error) {
	return model.GetAllTickets(pageNum, pageSize, status)
}

// GetTicketDetail retrieves ticket detail with permission check
func GetTicketDetail(ticketId int, userId int, isAdmin bool) (*model.Ticket, error) {
	ticket, err := model.GetTicketById(ticketId)
	if err != nil {
		return nil, err
	}
	
	// Permission check: user can only view their own tickets, admin can view all
	if !isAdmin && ticket.UserId != userId {
		return nil, errors.New("无权访问该工单")
	}
	
	return ticket, nil
}

// UpdateTicketStatus updates ticket status
func UpdateTicketStatus(ticketId int, status string, userId int, isAdmin bool) error {
	// Validate status
	if !common.ValidateTicketStatus(status) {
		return errors.New("无效的状态")
	}
	
	// Get ticket to check permissions
	ticket, err := model.GetTicketById(ticketId)
	if err != nil {
		return err
	}
	
	// Permission check
	if !isAdmin && ticket.UserId != userId {
		return errors.New("无权修改该工单")
	}
	
	return model.UpdateTicketStatus(ticketId, status)
}

// ReplyToTicket adds admin reply to ticket
func ReplyToTicket(ticketId int, reply string, adminId int) error {
	if reply == "" {
		return errors.New("回复内容不能为空")
	}
	
	// Sanitize reply
	reply = common.SanitizeInput(reply)
	
	// Update ticket with reply
	err := model.UpdateTicketReply(ticketId, reply, adminId)
	if err != nil {
		return err
	}
	
	// Update status to processing if it's pending
	ticket, err := model.GetTicketById(ticketId)
	if err == nil && ticket.Status == model.TicketStatusPending {
		model.UpdateTicketStatus(ticketId, model.TicketStatusProcessing)
	}
	
	return nil
}

// DeleteTicket soft deletes a ticket
func DeleteTicket(ticketId int, userId int, isAdmin bool) error {
	ticket, err := model.GetTicketById(ticketId)
	if err != nil {
		return err
	}
	
	// Permission check
	if !isAdmin && ticket.UserId != userId {
		return errors.New("无权删除该工单")
	}
	
	return ticket.Delete()
}
