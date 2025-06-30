package handler

import (
	"errors"
	"net/http"

	"github.com/1995parham-teaching/tinyurl/internal/domain/service/urlsvc"
	"github.com/1995parham-teaching/tinyurl/internal/infra/http/request"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// URL handles interaction with URLs.
type URL struct {
	Store  *urlsvc.URLSvc
	Logger *zap.Logger
	Tracer trace.Tracer
}

// Create generates short URL and save it on database.
// nolint: wrapcheck
func (h URL) Create(c echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.create")
	defer span.End()

	var rq request.URL

	err := c.Bind(&rq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = rq.Validate()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	span.SetAttributes(attribute.String("url", rq.URL))

	if rq.Name != "" {
		err := h.Store.CreateWithKey(ctx, rq.Name, rq.URL, rq.Expire)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			if errors.Is(err, urlsvc.ErrKeyAlreadyExists) {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}

			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusNoContent)
	}

	key, err := h.Store.Create(ctx, rq.URL, rq.Expire)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, key)
}

// Retrieve retrieves URL for given short URL and redirect to it.
// nolint: wrapcheck
func (h URL) Retrieve(c echo.Context) error {
	ctx, span := h.Tracer.Start(c.Request().Context(), "handler.url.retrieve")
	defer span.End()

	key := c.Param("key")

	url, err := h.Store.Visit(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.Redirect(http.StatusFound, url.URL)
}

// Register registers the routes of URL handler on given group.
func (h URL) Register(g *echo.Group) {
	g.GET("/:key", h.Retrieve)
	g.POST("/urls", h.Create)
}
