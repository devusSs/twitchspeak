package responses

// Custom error codes
var (
	CodeInternalError = "internal_error"
)

// Custom error messages
var (
	MessageInternalError = "Something went wrong, sorry about that."
)

// API / server response to indicate success
type Success struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// API / server response to indicate error
type Error struct {
	Code         int    `json:"code"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}
