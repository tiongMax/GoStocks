package alert

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/tiongMax/gostocks/proto/alert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the AlertService gRPC server.
type Server struct {
	pb.UnimplementedAlertServiceServer
	store *Store
}

// NewServer creates a new gRPC Alert Server with the given store.
func NewServer(store *Store) *Server {
	return &Server{store: store}
}

// CreateAlert creates a new price alert for a user.
func (s *Server) CreateAlert(ctx context.Context, req *pb.CreateAlertRequest) (*pb.CreateAlertResponse, error) {
	// Validate request
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id must be positive")
	}
	if req.Symbol == "" {
		return nil, status.Error(codes.InvalidArgument, "symbol is required")
	}
	if req.TargetPrice <= 0 {
		return nil, status.Error(codes.InvalidArgument, "target_price must be positive")
	}
	if req.Condition == pb.AlertCondition_CONDITION_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "condition must be ABOVE or BELOW")
	}

	// Convert proto condition to string
	condition := conditionToString(req.Condition)

	// Create alert in database
	alert, err := s.store.CreateAlert(
		int(req.UserId),
		strings.ToUpper(req.Symbol),
		req.TargetPrice,
		condition,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create alert: %v", err)
	}

	return &pb.CreateAlertResponse{
		AlertId: int32(alert.ID),
		Message: fmt.Sprintf("Alert created: %s %s $%.2f", alert.Symbol, condition, alert.TargetPrice),
	}, nil
}

// GetAlerts retrieves alerts based on filter criteria.
func (s *Server) GetAlerts(ctx context.Context, req *pb.GetAlertsRequest) (*pb.GetAlertsResponse, error) {
	// Fetch alerts from database
	alerts, err := s.store.GetAlertsByUser(int(req.UserId), req.ActiveOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get alerts: %v", err)
	}

	// Convert to proto messages
	pbAlerts := make([]*pb.Alert, len(alerts))
	for i, a := range alerts {
		pbAlerts[i] = &pb.Alert{
			Id:          int32(a.ID),
			UserId:      int32(a.UserID),
			Symbol:      a.Symbol,
			TargetPrice: a.TargetPrice,
			Condition:   stringToCondition(a.Condition),
			Triggered:   a.Triggered,
			CreatedAt:   a.CreatedAt.Unix(),
		}
	}

	return &pb.GetAlertsResponse{
		Alerts: pbAlerts,
	}, nil
}

// conditionToString converts proto AlertCondition to database string.
func conditionToString(c pb.AlertCondition) string {
	switch c {
	case pb.AlertCondition_ABOVE:
		return "ABOVE"
	case pb.AlertCondition_BELOW:
		return "BELOW"
	default:
		return ""
	}
}

// stringToCondition converts database string to proto AlertCondition.
func stringToCondition(s string) pb.AlertCondition {
	switch strings.ToUpper(s) {
	case "ABOVE":
		return pb.AlertCondition_ABOVE
	case "BELOW":
		return pb.AlertCondition_BELOW
	default:
		return pb.AlertCondition_CONDITION_UNSPECIFIED
	}
}
