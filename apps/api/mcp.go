package main

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/modules/mcp"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// mcpSetup is what the MCP server needs besides the services buildRouter builds. A nil
// setup leaves /mcp unmounted.
type mcpSetup struct {
	config mcp.Config
}

// newMCPSetup prepares the MCP server when MCP_RESOURCE is set and SSO is on, and nothing
// otherwise. The verifier runs Registre discovery here, so a wrong issuer fails at boot
// like the OIDC kit does. A person's first MCP call creates their account through the OIDC
// kit, as a first web login would.
func newMCPSetup(ctx context.Context, appEnv *env.Config, sqlDB *sql.DB, parts *authParts, appLogger *slog.Logger) (*mcpSetup, error) {
	if appEnv.OIDC == nil || appEnv.OIDC.MCPResource == "" {
		return nil, nil
	}
	verifier, err := mcpauth.Verifier(ctx, mcpauth.Config{
		Issuer:      appEnv.OIDC.Issuer,
		Resource:    appEnv.OIDC.MCPResource,
		ScopePrefix: "sablier",
		Identities:  parts.identities,
		Provision:   true,
		Accounts:    parts.kit,
		Logger:      appLogger,
	})
	if err != nil {
		return nil, err
	}
	appLogger.Info("mcp server on", slog.String("resource", appEnv.OIDC.MCPResource))
	return &mcpSetup{config: mcp.Config{
		Resource: appEnv.OIDC.MCPResource,
		Issuer:   appEnv.OIDC.Issuer,
		Version:  version,
		Origins:  appEnv.CORSAllowedOrigins,
		StateKey: mcpkit.StateKey("sablier", appEnv.OIDC.ClientSecret),
		Verifier: verifier,
		Pending:  mcpkit.PostgresPending{DB: sqlDB},
		Logger:   appLogger,
	}}, nil
}

// service builds the MCP server over deps, or nothing for a nil setup.
func (s *mcpSetup) service(deps mcp.Deps) (*mcp.Service, error) {
	if s == nil {
		return nil, nil
	}
	return mcp.NewService(deps, s.config)
}
