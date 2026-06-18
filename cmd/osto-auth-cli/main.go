package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"osto-auth-cli/internal/auth"
	"osto-auth-cli/internal/commands"
	"osto-auth-cli/internal/config"
	"osto-auth-cli/internal/db"
	"osto-auth-cli/internal/logger"
	"osto-auth-cli/internal/repl"
	"osto-auth-cli/internal/repository"
	"osto-auth-cli/internal/session"
	"osto-auth-cli/internal/state"
	"osto-auth-cli/internal/totp"
	"osto-auth-cli/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)

	dbConn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	if err := db.Migrate(dbConn, migrations.FS); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database initialized and migrations applied successfully")

	slog.Info("Starting osto-auth-cli",
		"db_path", cfg.DBPath,
		"session_time", cfg.SessionTime,
		"log_level", cfg.LogLevel.String(),
		"cleanup_interval", cfg.CleanupInterval,
		"lockout_threshold", cfg.LockoutThreshold,
		"lockout_duration", cfg.LockoutDuration,
	)

	appState := &state.AppState{}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "osto> ",
		HistoryFile:     "/tmp/osto_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		slog.Error("failed to initialize readline", "error", err)
		os.Exit(1)
	}
	defer rl.Close()

	userRepo := repository.NewSQLiteUserRepository(dbConn)
	sessionRepo := repository.NewSQLiteSessionRepository(dbConn)
	
	sessionService := session.NewSessionService(sessionRepo, cfg)
	totpService := totp.NewTOTPService()
	authService := auth.NewAuthService(userRepo, sessionService, totpService, cfg.AppEncryptionKey, cfg)
	authGuard := session.NewAuthGuard(sessionRepo, userRepo)
	
	enrollmentService := totp.NewEnrollmentService(userRepo, sessionService, totpService, cfg.AppEncryptionKey)

	preLogin := repl.NewRegistry()
	preLogin.Register(commands.NewRegisterCommand(rl, authService))
	preLogin.Register(commands.NewLoginCommand(rl, authService))
	preLogin.Register(commands.NewExitCommand())
	preLogin.Register(commands.NewHelpCommand(preLogin.All))

	postLogin := repl.NewRegistry()
	postLogin.Register(commands.NewWhoamiCommand(authGuard))
	postLogin.Register(commands.NewEnable2FACommand(rl, authGuard, totpService, enrollmentService))
	postLogin.Register(commands.NewDisable2FACommand(rl, authGuard, enrollmentService))
	postLogin.Register(commands.NewLogoutCommand(sessionService))
	postLogin.Register(commands.NewExitCommand())
	postLogin.Register(commands.NewHelpCommand(postLogin.All))

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				deleted, err := sessionRepo.DeleteExpired(ctx, time.Now())
				if err != nil {
					slog.Error("failed to cleanup expired sessions", "error", err)
				} else if deleted > 0 {
					slog.Debug("cleaned up expired sessions", "count", deleted)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	defer wg.Wait()
	defer cancel()

	fmt.Println("┌──────────────────────────────────────────────────┐")
	fmt.Println("│                                                  │")
	fmt.Println("│                   OSTO AUTH CLI                  │")
	fmt.Println("│         Secure REPL-based Identity Manager       │")
	fmt.Println("│                                                  │")
	fmt.Println("└──────────────────────────────────────────────────┘")

	r := repl.NewREPL(appState, preLogin, postLogin, rl)
	
	revoker := func() {
		if appState.IsAuthenticated() {
			_ = sessionService.Revoke(context.Background(), appState.SessionToken)
		}
	}
	
	if err := r.Run(revoker); err != nil {
		fmt.Fprintf(os.Stderr, "REPL exited with error: %v\n", err)
		os.Exit(1)
	}
}
