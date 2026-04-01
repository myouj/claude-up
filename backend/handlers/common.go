package handlers

import "prompt-vault/middleware"

// Re-exports from middleware for backward compatibility within the handlers package.
const (
	DefaultPage  = middleware.DefaultPage
	DefaultLimit = middleware.DefaultLimit
	MaxLimit     = middleware.MaxLimit
)
