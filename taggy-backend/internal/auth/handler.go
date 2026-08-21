package auth

import (
	"errors"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Handler struct {
	service      *Service
	log          zerolog.Logger
	exposeDevOTP bool
	frontendURL  string
}

func NewHandler(service *Service, log zerolog.Logger, exposeDevOTP bool, frontendURL string) *Handler {
	return &Handler{
		service:      service,
		log:          log,
		exposeDevOTP: exposeDevOTP,
		frontendURL:  strings.TrimRight(frontendURL, "/"),
	}
}

func (h *Handler) Register(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req registerRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid register payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	pending, otp, err := h.service.Register(c.Request().Context(), RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Name:     req.Name,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	log.Info().
		Str("email", pending.Email).
		Str("username", pending.Username).
		Msg("register handled")

	return c.JSON(http.StatusCreated, toRegisterResponse(pending, h.devOTP(otp)))
}

func (h *Handler) Login(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req loginRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid login payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userAgent := c.Request().UserAgent()
	var ip *netip.Addr
	if addr, err := netip.ParseAddr(c.RealIP()); err == nil {
		ip = &addr
	}

	pair, err := h.service.Login(c.Request().Context(), LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: &userAgent,
		IPAddress: ip,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			log.Warn().Str("email", req.Email).Msg("login failed")
		}
		return err
	}

	log.Info().Str("email", req.Email).Msg("login handled")
	return c.JSON(http.StatusOK, toTokenResponse(pair))
}

func (h *Handler) Refresh(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req refreshRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid refresh payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	pair, err := h.service.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) || errors.Is(err, ErrSessionExpired) {
			log.Warn().Msg("refresh failed")
		}
		return err
	}

	log.Info().Msg("refresh handled")
	return c.JSON(http.StatusOK, toTokenResponse(pair))
}

func (h *Handler) Logout(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req logoutRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid logout payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return err
	}

	log.Info().Msg("logout handled")
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) LogoutAll(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	claims, ok := ClaimsFromContext(c)
	if !ok {
		log.Warn().Msg("logout-all unauthorized")
		return apperrors.ErrUnauthorized
	}

	userPublicID, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Warn().Msg("logout-all invalid token subject")
		return apperrors.ErrUnauthorized
	}

	if err := h.service.LogoutAll(c.Request().Context(), userPublicID); err != nil {
		return err
	}

	log.Info().
		Str("user_id", userPublicID.String()).
		Msg("logout-all handled")

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) VerifyEmail(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req verifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	req.Otp = strings.TrimSpace(req.Otp)
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.service.VerifyEmail(c.Request().Context(), req.Email, req.Otp)
	if err != nil {
		return err
	}

	log.Info().Str("email", req.Email).Msg("verify email handled")
	return c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *Handler) ResendVerification(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req resendVerificationRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	otp, err := h.service.ResendVerification(c.Request().Context(), req.Email)
	if err != nil {
		return err
	}

	log.Info().Str("email", req.Email).Msg("resend verification handled")

	devOTP := h.devOTP(otp)
	if devOTP != "" {
		return c.JSON(http.StatusOK, resendVerificationResponse{DevOTP: devOTP})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GoogleStart(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	url, err := h.service.GoogleAuthURL()
	if err != nil {
		return err
	}

	log.Info().Msg("google oauth start handled")

	accept := c.Request().Header.Get(echo.HeaderAccept)
	wantsJSON := strings.Contains(accept, "application/json") &&
		!strings.Contains(accept, "text/html")
	if wantsJSON {
		return c.JSON(http.StatusOK, googleAuthURLResponse{URL: url})
	}

	return c.Redirect(http.StatusFound, url)
}

func (h *Handler) GoogleCallback(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if code == "" || state == "" {
		return h.redirectOAuthError(c, "missing code or state")
	}

	userAgent := c.Request().UserAgent()
	var ip *netip.Addr
	if addr, err := netip.ParseAddr(c.RealIP()); err == nil {
		ip = &addr
	}

	result, err := h.service.CompleteGoogleOAuth(c.Request().Context(), code, state, &userAgent, ip)
	if err != nil {
		log.Warn().Err(err).Msg("google oauth callback failed")
		return h.redirectOAuthError(c, err.Error())
	}

	wantsJSON := strings.Contains(c.Request().Header.Get("Accept"), "application/json") &&
		!strings.Contains(c.Request().Header.Get("Accept"), "text/html")

	if result.Pending != nil {
		log.Info().Str("email", result.Pending.Email).Msg("google oauth needs profile")
		if wantsJSON {
			return c.JSON(http.StatusOK, pendingGoogleRegistrationResponse{
				RegistrationToken: result.Pending.RegistrationToken,
				Email:             result.Pending.Email,
				Name:              result.Pending.Name,
				Picture:           result.Pending.Picture,
			})
		}
		q := url.Values{}
		q.Set("registration_token", result.Pending.RegistrationToken)
		q.Set("email", result.Pending.Email)
		q.Set("name", result.Pending.Name)
		if result.Pending.Picture != "" {
			q.Set("picture", result.Pending.Picture)
		}
		return c.Redirect(http.StatusFound, h.frontendURL+"/auth/complete-google#"+q.Encode())
	}

	log.Info().Msg("google oauth callback handled")
	if wantsJSON {
		return c.JSON(http.StatusOK, toTokenResponse(*result.Tokens))
	}

	q := url.Values{}
	q.Set("access_token", result.Tokens.AccessToken)
	q.Set("refresh_token", result.Tokens.RefreshToken)
	q.Set("username", result.Tokens.Username)
	return c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback#"+q.Encode())
}

func (h *Handler) CompleteGoogleRegistration(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	var req completeGoogleRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	userAgent := c.Request().UserAgent()
	var ip *netip.Addr
	if addr, err := netip.ParseAddr(c.RealIP()); err == nil {
		ip = &addr
	}

	pair, err := h.service.CompleteGoogleRegistration(
		c.Request().Context(),
		req.RegistrationToken,
		req.Username,
		req.Name,
		&userAgent,
		ip,
	)
	if err != nil {
		return err
	}

	log.Info().Str("username", pair.Username).Msg("google registration completed")
	return c.JSON(http.StatusCreated, toTokenResponse(pair))
}

func (h *Handler) redirectOAuthError(c echo.Context, message string) error {
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") &&
		!strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		return apperrors.ErrBadRequest
	}

	q := url.Values{}
	q.Set("error", message)
	return c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?"+q.Encode())
}

func (h *Handler) devOTP(otp string) string {
	if h.exposeDevOTP {
		return otp
	}
	return ""
}

func toUserResponse(user sqlc.User) userResponse {
	return userResponse{
		PublicID:      user.PublicID.String(),
		Email:         user.Email,
		Username:      user.Username,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
		Subscription:  string(user.Subscription),
		IsAdmin:       user.IsAdmin(),
	}
}

func toRegisterResponse(pending PendingSignup, devOTP string) registerResponse {
	return registerResponse{
		Email:         pending.Email,
		Username:      pending.Username,
		Name:          pending.Name,
		EmailVerified: false,
		Subscription:  string(sqlc.SubscriptionTierFREE),
		DevOTP:        devOTP,
	}
}

func toTokenResponse(pair TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		Username:     pair.Username,
		IsAdmin:      pair.IsAdmin,
		Subscription: pair.Subscription,
	}
}

func (h *Handler) AdminMe(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	publicID, err := UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	user, err := h.service.AdminMe(c.Request().Context(), publicID)
	if err != nil {
		return err
	}
	log.Info().Str("username", user.Username).Msg("admin me ok")
	return c.JSON(http.StatusOK, adminMeResponse{
		PublicID:     user.PublicID.String(),
		Username:     user.Username,
		IsAdmin:      user.IsAdmin(),
		Subscription: string(user.Subscription),
	})
}
