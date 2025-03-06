package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 標準APIレスポンスを表す構造体
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// エラー詳細を表す構造体
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// レスポンス用のメタデータを表す構造体
type MetaInfo struct {
	Total       int64 `json:"total,omitempty"`
	Count       int   `json:"count,omitempty"`
	Page        int   `json:"page,omitempty"`
	PerPage     int   `json:"per_page,omitempty"`
	TotalPages  int   `json:"total_pages,omitempty"`
	HasNext     bool  `json:"has_next,omitempty"`
	HasPrevious bool  `json:"has_previous,omitempty"`
}

// 成功レスポンスを作成する
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// エラーレスポンスを作成する
func NewErrorResponse(code, message string, details interface{}) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// ページネーション付きレスポンスを作成する
func NewPaginatedResponse(data interface{}, page, perPage int, total int64) Response {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	hasNext := page < totalPages
	hasPrevious := page > 1

	return Response{
		Success: true,
		Data:    data,
		Meta: &MetaInfo{
			Total:       total,
			Count:       len(data.([]interface{})),
			Page:        page,
			PerPage:     perPage,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}
}

// 指定したステータスコードでJSONレスポンスを送信する
func JSON(c *gin.Context, statusCode int, response Response) {
	c.JSON(statusCode, response)
}

// 成功レスポンスを送信する
func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, NewSuccessResponse(data))
}

// 作成成功レスポンスを送信する
func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, NewSuccessResponse(data))
}

// コンテンツなしレスポンスを送信する
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// 不正リクエストエラーレスポンスを送信する
func BadRequest(c *gin.Context, message string, details interface{}) {
	JSON(c, http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", message, details))
}

// 未認証エラーレスポンスを送信する
func Unauthorized(c *gin.Context, message string) {
	JSON(c, http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", message, nil))
}

// アクセス禁止エラーレスポンスを送信する
func Forbidden(c *gin.Context, message string) {
	JSON(c, http.StatusForbidden, NewErrorResponse("FORBIDDEN", message, nil))
}

// 見つからないエラーレスポンスを送信する
func NotFound(c *gin.Context, message string) {
	JSON(c, http.StatusNotFound, NewErrorResponse("NOT_FOUND", message, nil))
}

// 競合エラーレスポンスを送信する
func Conflict(c *gin.Context, message string, details interface{}) {
	JSON(c, http.StatusConflict, NewErrorResponse("CONFLICT", message, details))
}

// 内部サーバーエラーレスポンスを送信する
func InternalServerError(c *gin.Context, message string) {
	JSON(c, http.StatusInternalServerError, NewErrorResponse("INTERNAL_SERVER_ERROR", message, nil))
}

// バリデーションエラーレスポンスを送信する
func ValidationError(c *gin.Context, details interface{}) {
	JSON(c, http.StatusUnprocessableEntity, NewErrorResponse("VALIDATION_ERROR", "バリデーションに失敗しました", details))
}

// リクエスト過多エラーレスポンスを送信する
func TooManyRequests(c *gin.Context, message string) {
	JSON(c, http.StatusTooManyRequests, NewErrorResponse("TOO_MANY_REQUESTS", message, nil))
} 