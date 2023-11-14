package response

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

var (
	app *fiber.App
	c   *fiber.Ctx
)

func init() {
	app = fiber.New()
	c = app.AcquireCtx(&fasthttp.RequestCtx{})
}

func TestSuccessNoContent(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"TestSuccessNoContentValid", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SuccessNoContent(c); (err != nil) != tt.wantErr {
				t.Errorf("SuccessNoContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateResponse(t *testing.T) {
	type args struct {
		c          *fiber.Ctx
		statusCode int
		success    bool
		message    string
		data       interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Successful response with data",
			args: args{
				c:          c,
				statusCode: 200,
				success:    true,
				message:    "Success",
				data:       map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Error response with message",
			args: args{
				c:          c,
				statusCode: 400,
				success:    false,
				message:    "Bad Request",
				data:       nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateResponse(tt.args.c, tt.args.statusCode, tt.args.success, tt.args.message, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("CreateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuccessLoaded(t *testing.T) {
	type args struct {
		c    *fiber.Ctx
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Successful response with data",
			args: args{
				c:    c,
				data: map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Successful response with nil data",
			args: args{
				c:    c,
				data: nil,
			},
			wantErr: false,
		},
		// Add more test cases as needed.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SuccessLoaded(tt.args.c, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SuccessLoaded() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuccessCreated(t *testing.T) {
	type args struct {
		c    *fiber.Ctx
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Successful response with data",
			args: args{
				c:    c,
				data: map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Successful response with nil data",
			args: args{
				c:    c,
				data: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SuccessCreated(tt.args.c, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SuccessCreated() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBadRequest(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Bad Request with message",
			args: args{
				c:       c,
				message: "Bad Request",
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Bad Request with empty message",
			args: args{
				c:       c,
				message: "",
			},
			wantErr: false,
		},
		// Add more test cases as needed.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BadRequest(tt.args.c, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("BadRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnauthorized(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Unauthorized request",
			args: args{
				c: c,
			},
			wantErr: false,
		},
		// Add more test cases as needed.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unauthorized(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Unauthorized() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataNotFound(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Data not found with a message",
			args: args{
				c:       c,
				message: "Resource not found",
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Data not found with an empty message",
			args: args{
				c:       c,
				message: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DataNotFound(tt.args.c, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("DataNotFound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Custom error message",
			args: args{
				c:       c,
				message: "Something went wrong",
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Empty error message",
			args: args{
				c:       c,
				message: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Error(tt.args.c, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestErrorWithData(t *testing.T) {
	type args struct {
		c       *fiber.Ctx
		message string
		data    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Case 1: Custom error message with data",
			args: args{
				c:       c,
				message: "Something went wrong",
				data:    map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Test Case 2: Empty error message with data",
			args: args{
				c:       c,
				message: "",
				data:    map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Test Case 3: Custom error message with nil data",
			args: args{
				c:       c,
				message: "Error occurred",
				data:    nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrorWithData(tt.args.c, tt.args.message, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ErrorWithData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
