package gateway

import (
	"context"
	"fmt"
	"time"

	pb "github.com/tiongMax/gostocks/proto/alert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AlertClient wraps the gRPC connection to the Alert Service.
type AlertClient struct {
	conn   *grpc.ClientConn
	client pb.AlertServiceClient
}

// NewAlertClient creates a new gRPC client connection to the Alert Service.
func NewAlertClient(addr string) (*AlertClient, error) {
	// Set up connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Alert Service: %w", err)
	}

	client := pb.NewAlertServiceClient(conn)

	return &AlertClient{
		conn:   conn,
		client: client,
	}, nil
}

// CreateAlertRequest represents the request body for creating an alert.
type CreateAlertRequest struct {
	UserID      int32   `json:"user_id" binding:"required,gt=0"`
	Symbol      string  `json:"symbol" binding:"required"`
	TargetPrice float64 `json:"target_price" binding:"required,gt=0"`
	Condition   string  `json:"condition" binding:"required,oneof=ABOVE BELOW"`
}

// CreateAlertResponse represents the response after creating an alert.
type CreateAlertResponse struct {
	AlertID int32  `json:"alert_id"`
	Message string `json:"message"`
}

// CreateAlert creates a new alert via the Alert Service.
func (a *AlertClient) CreateAlert(ctx context.Context, req *CreateAlertRequest) (*CreateAlertResponse, error) {
	// Convert condition string to proto enum
	condition := pb.AlertCondition_CONDITION_UNSPECIFIED
	switch req.Condition {
	case "ABOVE":
		condition = pb.AlertCondition_ABOVE
	case "BELOW":
		condition = pb.AlertCondition_BELOW
	}

	// Call gRPC
	resp, err := a.client.CreateAlert(ctx, &pb.CreateAlertRequest{
		UserId:      req.UserID,
		Symbol:      req.Symbol,
		TargetPrice: req.TargetPrice,
		Condition:   condition,
	})
	if err != nil {
		return nil, err
	}

	return &CreateAlertResponse{
		AlertID: resp.AlertId,
		Message: resp.Message,
	}, nil
}

// AlertData represents a single alert in the response.
type AlertData struct {
	ID          int32   `json:"id"`
	UserID      int32   `json:"user_id"`
	Symbol      string  `json:"symbol"`
	TargetPrice float64 `json:"target_price"`
	Condition   string  `json:"condition"`
	Triggered   bool    `json:"triggered"`
	CreatedAt   int64   `json:"created_at"`
}

// GetAlerts retrieves alerts from the Alert Service.
func (a *AlertClient) GetAlerts(ctx context.Context, userID int32, activeOnly bool) ([]AlertData, error) {
	resp, err := a.client.GetAlerts(ctx, &pb.GetAlertsRequest{
		UserId:     userID,
		ActiveOnly: activeOnly,
	})
	if err != nil {
		return nil, err
	}

	alerts := make([]AlertData, len(resp.Alerts))
	for i, alert := range resp.Alerts {
		condition := "UNKNOWN"
		switch alert.Condition {
		case pb.AlertCondition_ABOVE:
			condition = "ABOVE"
		case pb.AlertCondition_BELOW:
			condition = "BELOW"
		}

		alerts[i] = AlertData{
			ID:          alert.Id,
			UserID:      alert.UserId,
			Symbol:      alert.Symbol,
			TargetPrice: alert.TargetPrice,
			Condition:   condition,
			Triggered:   alert.Triggered,
			CreatedAt:   alert.CreatedAt,
		}
	}

	return alerts, nil
}

// Close closes the gRPC connection.
func (a *AlertClient) Close() error {
	return a.conn.Close()
}

