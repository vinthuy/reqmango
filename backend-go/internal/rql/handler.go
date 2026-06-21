// backend-go/internal/rql/handler.go

package rql

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"gorm.io/gorm"
)

type RQLHandler struct {
	db          *gorm.DB
	rqlService *RQLService
}

func NewRQLHandler(db *gorm.DB) *RQLHandler {
	return &RQLHandler{
		db:          db,
		rqlService: NewRQLService(),
	}
}

func (h *RQLHandler) Search(c *gin.Context) {
	var req request.RQLSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 根据 entity 类型执行不同的查询
	switch req.Entity {
	case "issue":
		h.searchIssues(c, req)
	case "cycle":
		h.searchCycles(c, req)
	case "module":
		h.searchModules(c, req)
	default:
		c.JSON(http.StatusBadRequest, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "INVALID_ENTITY",
				Message: "不支持的实体类型: " + req.Entity,
			},
		})
	}
}

func (h *RQLHandler) searchIssues(c *gin.Context, req request.RQLSearchRequest) {
	// 如果 RQL 为空，返回所有工作项
	if req.RQL == "" {
		// TODO: 实现空查询的默认行为
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: true,
			Data: map[string]interface{}{
				"items":     []interface{}{},
				"total":     0,
				"page":      req.Page,
				"page_size": req.PageSize,
			},
		})
		return
	}

	// 验证 RQL 语法（只做验证，不执行）
	lexer := NewLexer(req.RQL)
	tokens, lexErr := lexer.Tokenize()
	if lexErr != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "RQL_LEX_ERROR",
				Message: lexErr.Error(),
			},
		})
		return
	}

	parser := NewParser(tokens)
	_, parseErr := parser.Parse()
	if parseErr != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "RQL_PARSE_ERROR",
				Message: parseErr.Error(),
			},
		})
		return
	}

	// 执行查询
	issues, total, err := h.rqlService.SearchIssues(h.db, req.ProjectID, req.RQL, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "QUERY_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, request.RQLSearchResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":     issues,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}

func (h *RQLHandler) searchCycles(c *gin.Context, req request.RQLSearchRequest) {
	cycles, total, err := h.rqlService.SearchCycles(h.db, req.ProjectID, req.RQL, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "QUERY_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, request.RQLSearchResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":     cycles,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}

func (h *RQLHandler) searchModules(c *gin.Context, req request.RQLSearchRequest) {
	modules, total, err := h.rqlService.SearchModules(h.db, req.ProjectID, req.RQL, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "QUERY_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, request.RQLSearchResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":     modules,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
		},
	})
}
