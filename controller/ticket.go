package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type CreateTicketRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Priority string `json:"priority" binding:"required"`
	Category string `json:"category" binding:"required"`
}

type ReplyTicketRequest struct {
	Reply string `json:"reply" binding:"required"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// CreateTicket creates a new ticket
func CreateTicket(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}
	
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	
	userIdInt := userId.(int)
	
	ticket, err := service.CreateTicket(userIdInt, req.Title, req.Content, req.Priority, req.Category)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工单创建成功",
		"data":    ticket,
	})
}

// GetTicketList retrieves ticket list for current user or all tickets for admin
func GetTicketList(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	
	userIdInt := userId.(int)
	role := session.Get("role")
	isAdmin := role != nil && (role.(int) >= common.RoleAdminUser)
	
	// Get pagination parameters
	pageNum, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	var tickets interface{}
	var total int64
	var err error
	
	if isAdmin {
		// Admin can see all tickets
		tickets, total, err = service.GetAllTicketsForAdmin(pageNum, pageSize, status)
	} else {
		// Regular user can only see their own tickets
		tickets, total, err = service.GetUserTickets(userIdInt, pageNum, pageSize)
	}
	
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    tickets,
		"total":   total,
	})
}

// GetTicketDetail retrieves ticket detail
func GetTicketDetail(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	
	userIdInt := userId.(int)
	role := session.Get("role")
	isAdmin := role != nil && (role.(int) >= common.RoleAdminUser)
	
	ticketId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的工单ID",
		})
		return
	}
	
	ticket, err := service.GetTicketDetail(ticketId, userIdInt, isAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    ticket,
	})
}

// ReplyTicket allows admin to reply to a ticket
func ReplyTicket(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	
	userIdInt := userId.(int)
	
	ticketId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的工单ID",
		})
		return
	}
	
	var req ReplyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}
	
	err = service.ReplyToTicket(ticketId, req.Reply, userIdInt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "回复成功",
	})
}

// UpdateTicketStatus updates ticket status
func UpdateTicketStatus(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	
	userIdInt := userId.(int)
	role := session.Get("role")
	isAdmin := role != nil && (role.(int) >= common.RoleAdminUser)
	
	ticketId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的工单ID",
		})
		return
	}
	
	var req UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求参数",
		})
		return
	}
	
	err = service.UpdateTicketStatus(ticketId, req.Status, userIdInt, isAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "状态更新成功",
	})
}

// DeleteTicket deletes a ticket
func DeleteTicket(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("id")
	if userId == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}
	
	userIdInt := userId.(int)
	role := session.Get("role")
	isAdmin := role != nil && (role.(int) >= common.RoleAdminUser)
	
	ticketId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的工单ID",
		})
		return
	}
	
	err = service.DeleteTicket(ticketId, userIdInt, isAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工单删除成功",
	})
}
