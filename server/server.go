package server

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/angelorc/mc-subscriber/config"
	_ "github.com/angelorc/mc-subscriber/swagger"
	"github.com/hanzoai/gochimp3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

type Server struct {
	*echo.Echo
	mc     *config.MailchimpConfig
	logger *zap.Logger
}

// @title BitSong -> Mailchimp subscriber
// @version 1.0
// @description The bitsong mailchimp subscriber proxy.

func NewServer(mc *config.MailchimpConfig, logger *zap.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Debug = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	s := &Server{
		Echo:   e,
		mc:     mc,
		logger: logger,
	}
	s.registerRoutes()

	return s
}

func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

func (s *Server) registerRoutes() {
	s.POST("/subscribe", s.PostSubscribe)
	s.GET("/swagger/*", echoSwagger.WrapHandler)
}

type PostResponse struct {
	Status string `json:"status"`
}

type PostRequest struct {
	Email  string `json:"email"`
	ListID string `json:"listID"`
}

// PostSubscribe godoc
// @Summary Subsrcibe email.
// @Description Subscribe an email address.
// @Accept json
// @Produce json
// @Success 200 {object} PostResponse
// @Param request body PostRequest true "Email address and ListID"
// @Router /subscribe [post]
func (s *Server) PostSubscribe(c echo.Context) error {
	pr := new(PostRequest)
	if err := c.Bind(pr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	email := strings.TrimSpace(pr.Email)
	listID := strings.TrimSpace(pr.ListID)

	// TODO: improve check
	ok := isValidEmail(email)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "email not valid")
	}

	if err := s.subscribeEmail(email, listID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, &PostResponse{Status: "ok"})
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (s *Server) subscribeEmail(email string, listID string) error {
	client := gochimp3.New(s.mc.APIKey)
	client.Timeout = 10 * time.Second

	req := &gochimp3.MemberRequest{
		EmailAddress: email,
		Status:       "pending",
	}

	list, err := client.GetList(listID, nil)
	if err != nil {
		return fmt.Errorf("failed to get list %s", listID)
	}

	if _, err := list.CreateMember(req); err != nil {
		return fmt.Errorf("failed to subscribe %s", req.EmailAddress)
	}

	return nil
}
