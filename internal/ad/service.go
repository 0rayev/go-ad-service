/*
This file encapsulates the business logic of the application.
It handles operations related to ads, such as validation, processing,
and coordination between different components (like the repository and the handler).
*/

package ad

import (
	"ad_service/pkg/cache"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var adCache = cache.NewCache()

type AdService struct {
	Repo *Repository
}

// AddAd adds a new ad to the database, with tracing
func (s *AdService) AddAd(ad *Ad, ctx context.Context) error {
	tracer := otel.Tracer("ad-service.service")
	ctx, span := tracer.Start(ctx, "AddAdService")
	defer span.End()

	err := s.Repo.AddAd(ad, ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to add ad")
		return err
	}

	span.SetAttributes(attribute.Int("ad_id", ad.ID), attribute.String("status", "success"))
	return nil
}

// GetAllAds retrieves ads from the database with pagination and sorting, with tracing
func (s *AdService) GetAllAds(page, limit int, sortBy, order string, ctx context.Context) ([]Ad, error) {
	tracer := otel.Tracer("ad-service.service")
	ctx, span := tracer.Start(ctx, "GetAllAdsService")
	defer span.End()

	ads, err := s.Repo.GetAllAds(page, limit, sortBy, order, ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve ads")
		return nil, err
	}

	span.SetAttributes(attribute.Int("ads_count", len(ads)), attribute.String("status", "success"))
	return ads, nil
}

// GetAdByID retrieves a single ad by its ID, with tracing and caching

func (s *AdService) GetAdByID(id int, ctx context.Context) (*Ad, error) {
	tracer := otel.Tracer("ad-service.service")
	ctx, span := tracer.Start(ctx, "GetAdByIDService")
	defer span.End()

	cacheKey := "ad_" + strconv.Itoa(id)

	// Trace cache retrieval attempt
	cachedAd, err := adCache.Get(cacheKey, ctx)
	if err == nil && cachedAd != "" {
		span.SetAttributes(attribute.String("cache_status", "found"), attribute.String("cache_key", cacheKey))

		var ad Ad
		if err := json.Unmarshal([]byte(cachedAd), &ad); err == nil {
			return &ad, nil
		}
		span.RecordError(err)
	} else {
		span.SetAttributes(attribute.String("cache_status", "not found"), attribute.String("cache_key", cacheKey))
	}

	// Cache miss, trace database query
	ad, err := s.Repo.GetAdByID(id, ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			span.SetAttributes(attribute.Int("ad_id", id), attribute.String("error", "Ad not found"))
			return nil, err // Return sql.ErrNoRows directly
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to retrieve ad")
		return nil, err
	}

	// Cache the result
	adBytes, err := json.Marshal(ad)
	if err == nil {
		adCache.Set(cacheKey, string(adBytes), 5*time.Minute, ctx)
		span.SetAttributes(attribute.String("cache_status", "set"))
	} else {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to set to cache")
	}

	span.SetAttributes(attribute.Int("ad_id", ad.ID), attribute.String("db_status", "Successfully retrieved by ID"))
	return ad, nil
}

// UpdateAd updates an existing ad, with tracing
func (s *AdService) UpdateAd(id int, ad *Ad, ctx context.Context) error {
	tracer := otel.Tracer("ad-service.service")
	ctx, span := tracer.Start(ctx, "UpdateAdService")
	defer span.End()

	err := s.Repo.UpdateAd(id, ad, ctx)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, ErrAdNotFound) {
			span.SetStatus(codes.Error, "Ad not found")
			return ErrAdNotFound
		}
		span.SetStatus(codes.Error, "Failed to update ad")
		return err
	}

	// Invalidate cache for this ad
	cacheKey := "ad_" + strconv.Itoa(id)
	adCache.Delete(cacheKey, ctx)

	span.SetAttributes(attribute.Int("ad_id", id), attribute.String("status", "updated"))
	return nil
}

// DeleteAd deletes an ad by ID, with tracing
func (s *AdService) DeleteAd(id int, ctx context.Context) error {
	tracer := otel.Tracer("ad-service.service")
	ctx, span := tracer.Start(ctx, "DeleteAdService")
	defer span.End()

	err := s.Repo.DeleteAd(id, ctx)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, ErrAdNotFound) {
			span.SetStatus(codes.Error, "Ad not found")
			return ErrAdNotFound
		}
		span.SetStatus(codes.Error, "Failed to delete ad")
		return err
	}

	// Invalidate cache for this ad
	cacheKey := "ad_" + strconv.Itoa(id)
	adCache.Delete(cacheKey, ctx)

	span.SetAttributes(attribute.Int("ad_id", id), attribute.String("status", "deleted"))
	return nil
}
