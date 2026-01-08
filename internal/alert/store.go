package alert

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(connStr string) (*Store, error) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate automatically migrates the database schema using GORM models.
func (s *Store) AutoMigrate() error {
	if err := s.db.AutoMigrate(&User{}, &Alert{}); err != nil {
		return fmt.Errorf("failed to auto migrate schema: %w", err)
	}
	return nil
}

// CreateUser creates a new user and returns their ID.
func (s *Store) CreateUser(username string) (int, error) {
	user := User{Username: username}
	if err := s.db.Create(&user).Error; err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	return user.ID, nil
}

// GetUserByUsername retrieves a user by their username.
func (s *Store) GetUserByUsername(username string) (*User, error) {
	var user User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// CreateAlert creates a new alert for a user.
func (s *Store) CreateAlert(userID int, symbol string, targetPrice float64, condition string) (*Alert, error) {
	alert := Alert{
		UserID:      userID,
		Symbol:      symbol,
		TargetPrice: targetPrice,
		Condition:   condition,
	}
	if err := s.db.Create(&alert).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}
	return &alert, nil
}

// GetActiveAlerts retrieves all untriggered alerts.
func (s *Store) GetActiveAlerts() ([]Alert, error) {
	var alerts []Alert
	if err := s.db.Where("triggered = ?", false).Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}
	return alerts, nil
}

// MarkAlertTriggered marks an alert as triggered.
func (s *Store) MarkAlertTriggered(alertID int) error {
	if err := s.db.Model(&Alert{}).Where("id = ?", alertID).Update("triggered", true).Error; err != nil {
		return fmt.Errorf("failed to mark alert as triggered: %w", err)
	}
	return nil
}

// GetAlertsByUser retrieves alerts for a specific user.
// If userID is 0, retrieves alerts for all users.
// If activeOnly is true, only non-triggered alerts are returned.
func (s *Store) GetAlertsByUser(userID int, activeOnly bool) ([]Alert, error) {
	var alerts []Alert
	query := s.db.Model(&Alert{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if activeOnly {
		query = query.Where("triggered = ?", false)
	}

	if err := query.Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	return alerts, nil
}

// GetActiveAlertsBySymbol retrieves all untriggered alerts for a specific symbol.
// This is optimized for the trigger logic to check only relevant alerts.
func (s *Store) GetActiveAlertsBySymbol(symbol string) ([]Alert, error) {
	var alerts []Alert
	if err := s.db.Where("symbol = ? AND triggered = ?", symbol, false).Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to query alerts for symbol %s: %w", symbol, err)
	}
	return alerts, nil
}
