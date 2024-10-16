/*
This file interacts with the database and handles data persistence.
It contains methods to perform CRUD operations, making it responsible for all database interactions.
*/
package ad

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Ad struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

type Repository struct {
	DB *sql.DB
}

// For returning Ad not found error, using in UpdateAd and DeleteAd
var ErrAdNotFound = errors.New("Ad not found")

// AddAd adds a new ad to the database, with tracing
func (r *Repository) AddAd(ad *Ad, ctx context.Context) error {
	tracer := otel.Tracer("ad-service.repository")
	ctx, span := tracer.Start(ctx, "AddAdRepository")
	defer span.End()
	// Build the SQL query
	query := "INSERT INTO ads (title, description, price, is_active) VALUES (?, ?, ?, ?)"

	result, err := r.DB.ExecContext(ctx, query, ad.Title, ad.Description, ad.Price, ad.IsActive)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to insert ad")
		return fmt.Errorf("could not insert ad: %v", err)
	}

	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve last insert ID")
		return fmt.Errorf("could not retrieve last insert ID: %v", err)
	}

	// Retrieve the created_at value from the database
	query = "SELECT created_at FROM ads WHERE id = ?"
	row := r.DB.QueryRowContext(ctx, query, id)
	var createdAt time.Time
	err = row.Scan(&createdAt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve created_at")
		return fmt.Errorf("could not retrieve created_at: %v", err)
	}

	// Update the Ad struct with the new ID and created_at time
	ad.ID = int(id)
	ad.CreatedAt = createdAt

	span.SetAttributes(attribute.Int("ad_id", ad.ID), attribute.String("status", "success"))
	return nil
}

// UpdateAd updates an existing ad, with tracing
func (r *Repository) UpdateAd(id int, ad *Ad, ctx context.Context) error {
	tracer := otel.Tracer("ad-service.repository")
	ctx, span := tracer.Start(ctx, "UpdateAdRepository")
	defer span.End()

	// Build the SQL query
	query := "UPDATE ads SET title = ?, description = ?, price = ?, "
	params := []interface{}{ad.Title, ad.Description, ad.Price}
	if ad.IsActive {
		query += "is_active = ?, "
		params = append(params, ad.IsActive)
	}
	query = query[:len(query)-2] // Remove last comma and space
	query += " WHERE id = ?"
	params = append(params, id)

	result, err := r.DB.ExecContext(ctx, query, params...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update ad")
		return fmt.Errorf("could not update ad: %v", err)
	}

	// Check if any rows were affected (if no rows, the ad wasn't found)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("could not retrieve affected rows: %v", err)
	}
	if rowsAffected == 0 {
		span.RecordError(ErrAdNotFound)
		span.SetStatus(codes.Error, "Ad not found")
		return ErrAdNotFound
	}

	span.SetAttributes(attribute.Int("ad_id", id), attribute.String("status", "updated"))
	return nil
}

// GetAllAds retrieves ads from the database with pagination and sorting, with tracing
func (r *Repository) GetAllAds(page, limit int, sortBy, order string, ctx context.Context) ([]Ad, error) {
	// Start a new tracing span for the GetAllAds operation
	tracer := otel.Tracer("ad-service.repository")
	ctx, span := tracer.Start(ctx, "GetAllAdsRepository")
	defer span.End()

	// Calculate the offset based on the current page and limit.
	offset := (page - 1) * limit
	query := fmt.Sprintf("SELECT id, title, description, price, created_at, is_active FROM ads ORDER BY %s %s LIMIT ? OFFSET ?", sortBy, order)

	rows, err := r.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve ads")
		return nil, err
	}
	defer rows.Close()

	// Go through the returned rows to get each ad
	ads := []Ad{}
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.ID, &ad.Title, &ad.Description, &ad.Price, &ad.CreatedAt, &ad.IsActive); err != nil {
			span.RecordError(err)
			return nil, err
		}
		ads = append(ads, ad)
	}

	span.SetAttributes(attribute.Int("ads_count", len(ads)), attribute.String("status", "success"))
	return ads, nil
}

// GetAdByID fetches the ad by its ID from the database, with tracing

func (r *Repository) GetAdByID(id int, ctx context.Context) (*Ad, error) {
	// Start a new tracing span for the GetAdByID operation
	tracer := otel.Tracer("ad-service.repository")
	ctx, span := tracer.Start(ctx, "GetAdByIDRepository")
	defer span.End()
	// Prepare the SQL query to select an ad by its ID
	query := "SELECT id, title, description, price, created_at, is_active FROM ads WHERE id = ?"
	var ad Ad
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&ad.ID, &ad.Title, &ad.Description, &ad.Price, &ad.CreatedAt, &ad.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			// No ad found with the given ID
			span.SetStatus(codes.Error, "Ad not found in DB")
			return nil, err
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to query ad from DB")
		return nil, err
	}

	span.SetAttributes(attribute.Int("ad_id", ad.ID), attribute.String("db_status", "success"))
	return &ad, nil
}

// DeleteAd deletes an ad by ID, with tracing
func (r *Repository) DeleteAd(id int, ctx context.Context) error {
	tracer := otel.Tracer("ad-service.repository")
	ctx, span := tracer.Start(ctx, "DeleteAdRepository")
	defer span.End()
	// Prepare the SQL query to delete the ad by its ID
	query := "DELETE FROM ads WHERE id = ?"
	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete ad")
		return fmt.Errorf("could not delete ad: %v", err)
	}

	// Check if any rows were affected (if no rows, the ad wasn't found)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("could not retrieve affected rows: %v", err)
	}
	if rowsAffected == 0 {
		span.RecordError(ErrAdNotFound)
		span.SetStatus(codes.Error, "Ad not found")
		return ErrAdNotFound // Ad not found
	}

	span.SetAttributes(attribute.Int("ad_id", id), attribute.String("status", "deleted"))
	return nil
}
