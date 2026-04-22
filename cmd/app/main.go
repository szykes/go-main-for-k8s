package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/subosito/gotenv"
	"github.com/szykes/go-main-for-k8s/app"
	"github.com/szykes/go-main-for-k8s/buildinfo"
	"github.com/szykes/go-main-for-k8s/logger"
)

type errorCode int

const (
	// errorCodeSuccess is an error code meaning the application finished without any issues.
	// Success is part of POSIX standard.
	errorCodeSuccess errorCode = 0
	// errorCodeGeneral is an error code meaning the application encountered an error during its ordinary execution.
	// General error is part of POSIX standard.
	errorCodeGeneral errorCode = 1
	// errorCodePanicOccured is an error code meaning the application encountered a panic during its execution.
	// Go runtime uses the same error code to indicate a panic occurred.
	errorCodePanicOccured errorCode = 2
	// errorCodeInvalidArgumentToExit is an exit code meaning invalid argument to exit
	// InvalidArgumentToExit is part of POSIX.
	// Don't use this directly because it is an offset of signals related error.
	errorCodeInvalidExitArgument errorCode = 128
)

func main() {
	code := new(errorCode)
	*code = errorCodeSuccess

	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic happened", "message", r, "call_trace", string(debug.Stack()))
			*code = errorCodePanicOccured
		}

		os.Exit(int(*code))
	}()

	*code = run()
}

func run() errorCode {
	logger.Init()

	isDebugModeEnabled := os.Getenv("DEBUG_MODE") == "true"

	buildinfo.Log(isDebugModeEnabled)

	envVarsFile := os.Getenv("ENV_VARS_FILE")
	if envVarsFile != "" {
		err := gotenv.Load(envVarsFile)
		if err != nil {
			slog.Error("Failed to load env file in main", "error", err)
			return errorCodeGeneral
		}
	} else {
		slog.Info("No env vars file is loaded")
	}

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	sigCh := make(chan os.Signal, 1)

	// !!! Do not handle the syscall.SIGKILL/os.Kill !!!
	// Set the `terminationGracePeriodSeconds` of pod/deployment in kubernetes.
	// Kubernetes sends SIGTERM to the container(s), when the kubernetes wants to terminates it/them.
	// If the `terminationGracePeriodSeconds` ends but the container still runs, kubernetes sends SIGKILL.
	// So, do not override the default SIGKILL handler because this is the last chance of kubernetes.
	// syscall.SIGINT is for local development, sent by Ctrl + C.
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		errCh <- app.Run(ctx, isDebugModeEnabled)
	}()

	select {
	case sig := <-sigCh:
		slog.Info("Received signal, initiating graceful shutdown", "signal", sig.String())
		stop()

		// Wait for app to actually finish its cleanup.
		<-errCh

		sysSig, ok := sig.(syscall.Signal)
		if !ok {
			slog.Error("Failed to convert signal type in main")
			return errorCodeGeneral
		}

		return errorCodeInvalidExitArgument + errorCode(sysSig)

	case err := <-errCh:
		if err != nil {
			slog.Error("Failed to run app in main", "error", err)
			return errorCodeGeneral
		}

		return errorCodeSuccess
	}
}
