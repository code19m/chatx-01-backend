# Token Revocation Implementation Plan

## Current State Analysis

### Current Implementation
- **JWT Generation**: Custom JWT implementation in `pkg/token/jwt.go`
- **Token Validation**: Only validates signature and expiration time
- **No Revocation**: Tokens are stateless - once issued, they're valid until expiration
- **Claims Structure**:
  ```go
  type Claims struct {
      UserID int
      Role   string
      Type   string  // "access" or "refresh"
      Exp    int64
      Iat    int64
  }
  ```

### Problems
1. **No way to revoke tokens** before expiration
2. **Deleted users** can still use their tokens
3. **Logged out users** tokens remain valid
4. **Compromised tokens** cannot be invalidated

## Proposed Solution: Redis-Based Token Store

### Architecture Overview

```
┌─────────────┐                    ┌──────────┐
│   Client    │                    │  Redis   │
└──────┬──────┘                    └────┬─────┘
       │                                │
       │ 1. Login                       │
       ├───────────────>                │
       │                                │
       │ 2. Generate JWT + Store in Redis
       │                  ├────────────>│
       │                                │
       │ 3. Return JWT                  │
       │<───────────────                │
       │                                │
       │ 4. Request with JWT            │
       ├───────────────>                │
       │                                │
       │ 5. Validate JWT                │
       │ 6. Check Redis  ├────────────>│
       │                 │<─────────────┤
       │ 7. Allow/Deny                  │
       │<───────────────                │
```

## Implementation Plan

### Phase 1: Redis Infrastructure Setup

#### 1.1 Create Redis Client Package
**File**: `pkg/redis/redis.go`

```go
package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type Client struct {
    rdb *redis.Client
}

type Config struct {
    Host     string
    Port     string
    Password string
    DB       int
}

func NewClient(cfg Config) (*Client, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
        Password: cfg.Password,
        DB:       cfg.DB,
    })

    // Test connection
    if err := rdb.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to redis: %w", err)
    }

    return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
    return c.rdb.Close()
}

// Token operations
func (c *Client) StoreToken(ctx context.Context, tokenID string, userID int, ttl time.Duration) error
func (c *Client) TokenExists(ctx context.Context, tokenID string) (bool, error)
func (c *Client) RevokeToken(ctx context.Context, tokenID string) error
func (c *Client) RevokeAllUserTokens(ctx context.Context, userID int) error
func (c *Client) GetUserIDByToken(ctx context.Context, tokenID string) (int, error)
```

**Dependencies**:
- Add `github.com/redis/go-redis/v9` to go.mod

#### 1.2 Add Redis Configuration
**File**: `internal/config/config.go`

```go
type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

// In Load() function:
Redis: RedisConfig{
    Host:     getEnv("REDIS_HOST", "localhost"),
    Port:     getEnv("REDIS_PORT", "6379"),
    Password: getEnv("REDIS_PASSWORD", ""),
    DB:       getEnvInt("REDIS_DB", 0),
},
```

**Environment Variables**:
```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Phase 2: Token Storage Strategy

#### 2.1 Redis Key Design

```
Access Tokens:
  Key:   "token:access:{token_id}"
  Value: "{user_id}"
  TTL:   15 minutes (access token TTL)

Refresh Tokens:
  Key:   "token:refresh:{token_id}"
  Value: "{user_id}"
  TTL:   24 hours (refresh token TTL)

User Token Index:
  Key:   "user:tokens:{user_id}"
  Value: Set of token IDs
  TTL:   None (manually managed)
```

#### 2.2 Update JWT Claims
**File**: `pkg/token/jwt.go`

```go
type Claims struct {
    JTI    string `json:"jti"`     // JWT ID - unique token identifier
    UserID int    `json:"user_id"`
    Role   string `json:"role"`
    Type   string `json:"type"`
    Exp    int64  `json:"exp"`
    Iat    int64  `json:"iat"`
}
```

Generate unique JTI using UUID or similar:
```go
import "github.com/google/uuid"

jti := uuid.New().String()
```

### Phase 3: Token Service Layer

#### 3.1 Create Token Service
**File**: `pkg/token/service.go`

```go
package token

import (
    "context"
    "time"
)

type TokenStore interface {
    StoreToken(ctx context.Context, tokenID string, userID int, ttl time.Duration) error
    TokenExists(ctx context.Context, tokenID string) (bool, error)
    RevokeToken(ctx context.Context, tokenID string) error
    RevokeAllUserTokens(ctx context.Context, userID int) error
}

type Service struct {
    generator  Generator
    tokenStore TokenStore
}

func NewService(generator Generator, tokenStore TokenStore) *Service {
    return &Service{
        generator:  generator,
        tokenStore: tokenStore,
    }
}

// GenerateAndStore generates a token and stores it in Redis
func (s *Service) GenerateAndStore(ctx context.Context, userID int, role string, tokenType TokenType) (string, error)

// ValidateAndCheck validates JWT and checks Redis
func (s *Service) ValidateAndCheck(ctx context.Context, tokenString string) (*Claims, error)

// Revoke revokes a specific token
func (s *Service) Revoke(ctx context.Context, tokenString string) error

// RevokeAllUserTokens revokes all tokens for a user
func (s *Service) RevokeAllUserTokens(ctx context.Context, userID int) error
```

### Phase 4: Update Authentication Flow

#### 4.1 Update Login (Token Generation)
**File**: `internal/auth/usecase/authuc/usecase.go`

```go
func (uc *useCase) Login(ctx context.Context, req LoginReq) (*LoginResp, error) {
    // ... existing user validation ...

    // Generate and store access token
    accessToken, err := uc.tokenService.GenerateAndStore(
        ctx,
        user.ID,
        user.Role.String(),
        token.TokenTypeAccess,
    )
    if err != nil {
        return nil, errs.Wrap(op, err)
    }

    // Generate and store refresh token
    refreshToken, err := uc.tokenService.GenerateAndStore(
        ctx,
        user.ID,
        user.Role.String(),
        token.TokenTypeRefresh,
    )
    if err != nil {
        return nil, errs.Wrap(op, err)
    }

    // ... rest of the code ...
}
```

#### 4.2 Update Logout (Token Revocation)
**File**: `internal/auth/usecase/authuc/usecase.go`

```go
func (uc *useCase) Logout(ctx context.Context, req LogoutReq) error {
    const op = "authuc.Logout"

    // Extract token from request (passed from handler)
    err := uc.tokenService.Revoke(ctx, req.AccessToken)
    if err != nil {
        return errs.Wrap(op, err)
    }

    // Optionally revoke refresh token too
    if req.RefreshToken != "" {
        err = uc.tokenService.Revoke(ctx, req.RefreshToken)
        if err != nil {
            // Log but don't fail - access token already revoked
            slog.Warn("failed to revoke refresh token", "error", err)
        }
    }

    return nil
}
```

#### 4.3 Update Authentication Middleware
**File**: `internal/auth/portal/portal.go`

```go
type Portal struct {
    userRepo     domain.UserRepository
    tokenService *token.Service  // Changed from tokenGenerator
}

func (p *Portal) authenticate(r *http.Request) (auth.AuthenticatedUser, error) {
    // ... extract token ...

    // Validate JWT and check Redis
    claims, err := p.tokenService.ValidateAndCheck(r.Context(), tokenString)
    if err != nil {
        return au, errors.New("unauthorized: invalid or revoked token")
    }

    // ... rest of validation ...
}
```

### Phase 5: User Deletion Token Revocation

#### 5.1 Update DeleteUser Use Case
**File**: `internal/auth/usecase/useruc/usecase.go`

```go
func (uc *useCase) DeleteUser(ctx context.Context, req DeleteUserReq) error {
    const op = "useruc.DeleteUser"

    // Get user to ensure they exist
    _, err := uc.userRepo.GetByID(ctx, req.UserID)
    if err != nil {
        return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("user_id", "user not found"))
    }

    // Revoke all user tokens BEFORE deleting
    err = uc.tokenService.RevokeAllUserTokens(ctx, req.UserID)
    if err != nil {
        slog.Error("failed to revoke user tokens", "user_id", req.UserID, "error", err)
        // Don't fail deletion - continue
    }

    // Delete user
    err = uc.userRepo.Delete(ctx, req.UserID)
    if err != nil {
        return errs.Wrap(op, err)
    }

    return nil
}
```

### Phase 6: Additional Features

#### 6.1 Token Refresh with Revocation Check
```go
func (uc *useCase) RefreshToken(ctx context.Context, req RefreshTokenReq) (*RefreshTokenResp, error) {
    // Validate refresh token and check Redis
    claims, err := uc.tokenService.ValidateAndCheck(ctx, req.RefreshToken)
    if err != nil {
        return nil, errs.Wrap(op, errs.NewUnauthorizedError("invalid or revoked refresh token"))
    }

    // Revoke old refresh token
    _ = uc.tokenService.Revoke(ctx, req.RefreshToken)

    // Generate new access and refresh tokens
    // ...
}
```

#### 6.2 Admin Endpoint to Revoke User Tokens
**New Endpoint**: `DELETE /auth/users/{user_id}/tokens`

```go
func (c *ctrl) revokeUserTokens(w http.ResponseWriter, r *http.Request) {
    req, err := httptools.BindRequest[useruc.RevokeUserTokensReq](r)
    if err != nil {
        httptools.HandleError(w, err)
        return
    }

    err = c.userUsecase.RevokeUserTokens(r.Context(), req)
    if err != nil {
        httptools.HandleError(w, err)
        return
    }

    httptools.WriteResponse(http.StatusNoContent, w, nil)
}
```

## Implementation Steps (Ordered)

### Step 1: Setup Redis Infrastructure
1. Add Redis client dependency: `go get github.com/redis/go-redis/v9`
2. Create `pkg/redis/redis.go` with token operations
3. Add Redis config to `internal/config/config.go`
4. Update `.env.example` with Redis variables
5. Initialize Redis client in `app.go`

### Step 2: Update Token Package
1. Add `JTI` field to `Claims` struct
2. Generate unique JTI in `Generate()` method
3. Create `token.Service` with Redis integration
4. Implement `GenerateAndStore()`, `ValidateAndCheck()`, `Revoke()`, `RevokeAllUserTokens()`

### Step 3: Update Auth Use Cases
1. Inject `token.Service` instead of `token.Generator`
2. Update `Login()` to store tokens in Redis
3. Update `Logout()` to revoke tokens
4. Add token parameter to logout request

### Step 4: Update Auth Middleware
1. Update `Portal` to use `token.Service`
2. Change `authenticate()` to check Redis after JWT validation

### Step 5: Update User Management
1. Update `DeleteUser()` to revoke all user tokens
2. Inject `token.Service` into `useruc`
3. Add `RevokeUserTokens()` use case (optional admin feature)

### Step 6: Testing
1. Test token generation and storage
2. Test token validation with Redis check
3. Test logout token revocation
4. Test user deletion token revocation
5. Test expired tokens are not in Redis

## Migration Considerations

### Backward Compatibility
- **Existing tokens**: Old tokens without JTI will fail Redis check
- **Solution**: Implement grace period with fallback to JWT-only validation
- **Alternative**: Force all users to re-login after deployment

### Grace Period Implementation
```go
func (s *Service) ValidateAndCheck(ctx context.Context, tokenString string) (*Claims, error) {
    claims, err := s.generator.Validate(tokenString)
    if err != nil {
        return nil, err
    }

    // If no JTI (old token), check if grace period is active
    if claims.JTI == "" {
        if time.Now().Before(gracePeriodEnd) {
            // Allow old tokens during grace period
            return claims, nil
        }
        return nil, errors.New("token format outdated")
    }

    // Check Redis for new tokens
    exists, err := s.tokenStore.TokenExists(ctx, claims.JTI)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, errors.New("token has been revoked")
    }

    return claims, nil
}
```

## Performance Considerations

### Redis Performance
- **Read latency**: ~1ms for token lookup
- **Connection pooling**: Use Redis connection pool
- **Failure handling**: If Redis is down, deny all requests (fail-closed)

### Alternative: Fail-Open Strategy
```go
exists, err := s.tokenStore.TokenExists(ctx, claims.JTI)
if err != nil {
    // Log error and allow request (risky but maintains availability)
    slog.Error("redis check failed, allowing request", "error", err)
    return claims, nil
}
```

## Security Considerations

1. **Token Rotation**: Implement refresh token rotation
2. **TTL Alignment**: Ensure Redis TTL matches JWT exp claim
3. **Atomic Operations**: Use Redis transactions for multi-key operations
4. **Monitoring**: Add metrics for revoked token access attempts

## Estimated Implementation Time

- **Phase 1**: Redis Setup - 2 hours
- **Phase 2**: Token Storage - 2 hours
- **Phase 3**: Token Service - 3 hours
- **Phase 4**: Auth Flow Updates - 3 hours
- **Phase 5**: User Deletion - 1 hour
- **Phase 6**: Additional Features - 2 hours
- **Testing**: 3 hours

**Total**: ~16 hours

## Questions to Consider

1. **Grace period**: Do you want a grace period for existing tokens?
2. **Fail strategy**: Fail-open or fail-closed if Redis is unavailable?
3. **Token rotation**: Implement refresh token rotation?
4. **Admin features**: Should admins be able to manually revoke user tokens?
5. **Logging**: Should we log all token revocation events?
