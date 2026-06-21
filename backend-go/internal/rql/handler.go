// backend-go/internal/rql/handler.go

package rql

import (
	"net/http"

	"github.com/reqmanpy/backend-go/internal/dto/request"

	"github.com/gin-gonic/gin"
)

type RQLHandler struct{}

func NewRQLHandler() *RQLHandler {
	return &RQLHandler{}
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

	// 词法分析
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

	// 语法分析
	parser := NewParser(tokens)
	ast, parseErr := parser.Parse()
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

	// 构建查询
	builder := NewQueryBuilder()
	whereClause, args, buildErr := builder.Build(ast)
	if buildErr != nil {
		c.JSON(http.StatusOK, request.RQLSearchResponse{
			Success: false,
			Error: &request.RQLError{
				Code:    "RQL_BUILD_ERROR",
				Message: buildErr.Error(),
			},
		})
		return
	}

	// TODO: 根据 entity 类型执行不同的查询
	// 这里先返回模拟数据，实际需要接入数据库

	c.JSON(http.StatusOK, request.RQLSearchResponse{
		Success: true,
		Data: map[string]interface{}{
			"items":     []interface{}{},
			"total":     0,
			"page":      req.Page,
			"page_size": req.PageSize,
			"where":     whereClause,
			"args":      args,
		},
	})
}
