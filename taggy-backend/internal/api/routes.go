package api

import (
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/audio"
	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	"github.com/HR-Shekhar/taggy-backend/internal/billing"
	"github.com/HR-Shekhar/taggy-backend/internal/community"
	"github.com/HR-Shekhar/taggy-backend/internal/notification"
	"github.com/HR-Shekhar/taggy-backend/internal/pod"
	"github.com/HR-Shekhar/taggy-backend/internal/progress"
	"github.com/HR-Shekhar/taggy-backend/internal/quiz"
	"github.com/HR-Shekhar/taggy-backend/internal/report"
	"github.com/HR-Shekhar/taggy-backend/internal/roadmap"
	"github.com/HR-Shekhar/taggy-backend/internal/roadmaprequest"
	"github.com/HR-Shekhar/taggy-backend/internal/search"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/middleware"
	"github.com/HR-Shekhar/taggy-backend/internal/skill"
	"github.com/HR-Shekhar/taggy-backend/internal/skillrequest"
	"github.com/HR-Shekhar/taggy-backend/internal/user"
	"github.com/labstack/echo/v4"
)

type Routes struct {
	Auth           AuthRoutes
	User           UserRoutes
	Skill          SkillRoutes
	SkillRequest   SkillRequestRoutes
	Roadmap        RoadmapRoutes
	RoadmapRequest  RoadmapRequestRoutes
	Progress       ProgressRoutes
	Pod            PodRoutes
	Community      CommunityRoutes
	Audio          AudioRoutes
	Notification   NotificationRoutes
	Report         ReportRoutes
	Search         SearchRoutes
	Admin          AdminRoutes
	Billing        BillingRoutes
	Quiz           QuizRoutes
}

type AuthRoutes struct {
	Handler       *auth.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type UserRoutes struct {
	Handler               *user.Handler
	JWTMiddleware         echo.MiddlewareFunc
	OptionalJWTMiddleware echo.MiddlewareFunc
}

type SkillRoutes struct {
	Handler       *skill.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type SkillRequestRoutes struct {
	Handler       *skillrequest.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type RoadmapRoutes struct {
	Handler       *roadmap.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type RoadmapRequestRoutes struct {
	Handler       *roadmaprequest.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type ProgressRoutes struct {
	Handler       *progress.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type PodRoutes struct {
	Handler       *pod.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type CommunityRoutes struct {
	Handler       *community.Handler
	JWTMiddleware echo.MiddlewareFunc
	Hub           *community.Hub
}

type AudioRoutes struct {
	Handler       *audio.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type NotificationRoutes struct {
	Handler       *notification.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type ReportRoutes struct {
	Handler       *report.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type SearchRoutes struct {
	Handler       *search.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type AdminRoutes struct {
	JWTMiddleware   echo.MiddlewareFunc
	AdminMiddleware echo.MiddlewareFunc
	AuthHandler     *auth.Handler
	SkillRequest    *skillrequest.Handler
	RoadmapRequest  *roadmaprequest.Handler
}

type BillingRoutes struct {
	Handler       *billing.Handler
	JWTMiddleware echo.MiddlewareFunc
}

type QuizRoutes struct {
	Handler       *quiz.Handler
	JWTMiddleware echo.MiddlewareFunc
}

func RegisterRoutes(e *echo.Echo, routes Routes) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	// In-memory per-IP limits (single-process). Redis can replace this later.
	limitRegister := middleware.IPRateLimit(5, time.Minute)
	limitLogin := middleware.IPRateLimit(20, time.Minute)
	limitOTP := middleware.IPRateLimit(5, time.Minute)
	limitJoin := middleware.IPRateLimit(30, time.Minute)
	limitReport := middleware.IPRateLimit(10, time.Minute)
	limitCatalogRequest := middleware.IPRateLimit(5, time.Hour)
	limitCheckout := middleware.IPRateLimit(10, time.Minute)

	authGroup := e.Group("/auth")
	authGroup.POST("/register", routes.Auth.Handler.Register, limitRegister)
	authGroup.POST("/login", routes.Auth.Handler.Login, limitLogin)
	authGroup.POST("/refresh", routes.Auth.Handler.Refresh)
	authGroup.POST("/logout", routes.Auth.Handler.Logout)
	authGroup.POST("/logout-all", routes.Auth.Handler.LogoutAll, routes.Auth.JWTMiddleware)
	authGroup.POST("/verify-email", routes.Auth.Handler.VerifyEmail, limitOTP)
	authGroup.POST("/resend-verification", routes.Auth.Handler.ResendVerification, limitOTP)
	authGroup.GET("/google/start", routes.Auth.Handler.GoogleStart)
	authGroup.GET("/google/callback", routes.Auth.Handler.GoogleCallback)
	authGroup.POST("/google/complete", routes.Auth.Handler.CompleteGoogleRegistration)

	profileRead := e.Group("/users", routes.User.OptionalJWTMiddleware)
	profileRead.GET("/:username", routes.User.Handler.GetProfile)

	userGroup := e.Group("/users", routes.User.JWTMiddleware)
	userGroup.PATCH("/:username", routes.User.Handler.UpdateProfile)
	userGroup.POST("/:username/avatar", routes.User.Handler.UploadAvatar)
	userGroup.GET("/:username/skills", routes.Skill.Handler.ListMySkills)
	userGroup.GET("/:username/skills/:skillSlug/milestones", routes.Skill.Handler.ListMilestones)
	userGroup.PATCH("/:username/skills/:skillSlug/milestones/:milestoneSlug", routes.Skill.Handler.UpdateMilestone)
	userGroup.PUT("/:username/skills/:skillSlug/roadmap-version", routes.Skill.Handler.SwitchRoadmapVersion)
	userGroup.POST("/:username/study-sessions", routes.Progress.Handler.LogStudySession)
	userGroup.GET("/:username/study-sessions", routes.Progress.Handler.ListStudySessions)
	userGroup.GET("/:username/streak", routes.Progress.Handler.GetStreak)
	userGroup.GET("/:username/progress/summary", routes.Progress.Handler.GetProgressSummary)
	userGroup.GET("/:username/pods", routes.Pod.Handler.ListMyPods)
	userGroup.GET("/:username/notifications", routes.Notification.Handler.List)
	userGroup.POST("/:username/notifications/read-all", routes.Notification.Handler.MarkAllRead)
	userGroup.POST("/:username/notifications/clear-read", routes.Notification.Handler.ClearRead)
	userGroup.PATCH("/:username/notifications/:id/read", routes.Notification.Handler.MarkRead)
	userGroup.GET("/:username/reports", routes.Report.Handler.ListMine)
	userGroup.GET("/:username/skill-requests", routes.SkillRequest.Handler.ListMine)
	userGroup.POST("/:username/skill-requests/:id/cancel", routes.SkillRequest.Handler.Cancel)
	userGroup.GET("/:username/roadmap-edit-requests", routes.RoadmapRequest.Handler.ListMine)
	userGroup.POST("/:username/roadmap-edit-requests/:id/cancel", routes.RoadmapRequest.Handler.Cancel)

	skillGroup := e.Group("/skills", routes.Skill.JWTMiddleware)
	skillGroup.GET("", routes.Skill.Handler.ListSkills)
	skillGroup.GET("/similar", routes.SkillRequest.Handler.ListSimilar)
	skillGroup.POST("/requests", routes.SkillRequest.Handler.Create, limitCatalogRequest)
	skillGroup.GET("/:slug", routes.Skill.Handler.GetSkillBySlug)
	skillGroup.POST("/:slug/join", routes.Skill.Handler.JoinSkill, limitJoin)
	skillGroup.GET("/:slug/roadmap", routes.Roadmap.Handler.GetRoadmap)
	skillGroup.GET("/:slug/roadmap/versions", routes.Roadmap.Handler.ListVersions)
	skillGroup.GET("/:slug/roadmap/versions/:versionNumber", routes.Roadmap.Handler.GetVersion)
	skillGroup.POST("/:skillSlug/roadmap-edit-requests", routes.RoadmapRequest.Handler.Create, limitCatalogRequest)
	skillGroup.POST("/:skillSlug/pods", routes.Pod.Handler.CreatePod)
	skillGroup.GET("/:skillSlug/pods", routes.Pod.Handler.ListPodsBySkill)
	skillGroup.GET("/:skillSlug/community", routes.Community.Handler.GetCommunity)
	skillGroup.GET("/:skillSlug/community/channels", routes.Community.Handler.ListChannels)
	skillGroup.GET("/:skillSlug/community/channels/:channelSlug/messages", routes.Community.Handler.ListChannelMessages)
	skillGroup.POST("/:skillSlug/community/channels/:channelSlug/messages", routes.Community.Handler.SendChannelMessage)
	skillGroup.GET("/:skillSlug/community/channels/:channelSlug/audio-rooms", routes.Audio.Handler.ListChannelRooms)
	skillGroup.POST("/:skillSlug/community/channels/:channelSlug/audio-rooms", routes.Audio.Handler.CreateChannelRoom)
	skillGroup.GET("/:skillSlug/community/leaderboard", routes.Quiz.Handler.CommunityLeaderboard)

	podGroup := e.Group("/pods", routes.Pod.JWTMiddleware)
	podGroup.GET("/:podSlug", routes.Pod.Handler.GetPod)
	podGroup.DELETE("/:podSlug", routes.Pod.Handler.DeletePod)
	podGroup.POST("/:podSlug/join", routes.Pod.Handler.JoinPod, limitJoin)
	podGroup.POST("/:podSlug/leave", routes.Pod.Handler.LeavePod)
	podGroup.GET("/:podSlug/messages", routes.Community.Handler.ListPodMessages)
	podGroup.POST("/:podSlug/messages", routes.Community.Handler.SendPodMessage)
	podGroup.GET("/:podSlug/audio-rooms", routes.Audio.Handler.ListPodRooms)
	podGroup.POST("/:podSlug/audio-rooms", routes.Audio.Handler.CreatePodRoom)
	podGroup.POST("/:podSlug/members/:username/accept", routes.Pod.Handler.AcceptMember)
	podGroup.POST("/:podSlug/members/:username/reject", routes.Pod.Handler.RejectMember)
	podGroup.POST("/:podSlug/members/:username/remove", routes.Pod.Handler.RemoveMember)
	podGroup.POST("/:podSlug/members/:username/role", routes.Pod.Handler.SetMemberRole)
	podGroup.POST("/:podSlug/quizzes", routes.Quiz.Handler.Start)
	podGroup.GET("/:podSlug/quizzes/mine", routes.Quiz.Handler.ListMine)
	podGroup.GET("/:podSlug/quizzes/:id", routes.Quiz.Handler.Get)
	podGroup.POST("/:podSlug/quizzes/:id/complete", routes.Quiz.Handler.Complete)
	podGroup.POST("/:podSlug/quizzes/:id/questions/:order/start", routes.Quiz.Handler.StartQuestion)
	podGroup.POST("/:podSlug/quizzes/:id/questions/:order/answer", routes.Quiz.Handler.AnswerQuestion)
	podGroup.GET("/:podSlug/leaderboard", routes.Quiz.Handler.PodLeaderboard)

	messageGroup := e.Group("/messages", routes.Community.JWTMiddleware)
	messageGroup.PATCH("/:id", routes.Community.Handler.EditMessage)
	messageGroup.DELETE("/:id", routes.Community.Handler.DeleteMessage)

	// Live chat fan-out (JWT via ?token=). Room: pod:{slug} | channel:{skill}:{channel}
	e.GET("/ws/chat", routes.Community.Handler.ServeChatWS)

	audioGroup := e.Group("/audio-rooms", routes.Audio.JWTMiddleware)
	audioGroup.GET("/:roomId", routes.Audio.Handler.GetRoom)
	audioGroup.POST("/:roomId/join", routes.Audio.Handler.JoinRoom)
	audioGroup.POST("/:roomId/leave", routes.Audio.Handler.LeaveRoom)
	audioGroup.POST("/:roomId/end", routes.Audio.Handler.EndRoom)

	reportGroup := e.Group("/reports", routes.Report.JWTMiddleware)
	reportGroup.POST("", routes.Report.Handler.Create, limitReport)

	searchGroup := e.Group("/search", routes.Search.JWTMiddleware)
	searchGroup.GET("", routes.Search.Handler.Search)

	billingGroup := e.Group("/billing", routes.Billing.JWTMiddleware)
	billingGroup.GET("/status", routes.Billing.Handler.Status)
	billingGroup.POST("/checkout/order", routes.Billing.Handler.CreateOrder, limitCheckout)
	billingGroup.POST("/checkout/verify", routes.Billing.Handler.Verify, limitCheckout)
	e.POST("/billing/webhooks/razorpay", routes.Billing.Handler.Webhook)

	adminGroup := e.Group("/admin", routes.Admin.JWTMiddleware, routes.Admin.AdminMiddleware)
	adminGroup.GET("/me", routes.Admin.AuthHandler.AdminMe)
	adminGroup.GET("/skill-requests", routes.Admin.SkillRequest.AdminList)
	adminGroup.POST("/skill-requests/:id/approve", routes.Admin.SkillRequest.AdminApprove)
	adminGroup.POST("/skill-requests/:id/reject", routes.Admin.SkillRequest.AdminReject)
	adminGroup.GET("/roadmap-edit-requests", routes.Admin.RoadmapRequest.AdminList)
	adminGroup.POST("/roadmap-edit-requests/:id/approve", routes.Admin.RoadmapRequest.AdminApprove)
	adminGroup.POST("/roadmap-edit-requests/:id/reject", routes.Admin.RoadmapRequest.AdminReject)
}
