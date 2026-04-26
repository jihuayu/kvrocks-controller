/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/apache/kvrocks-controller/config"
	"github.com/apache/kvrocks-controller/consts"
	"github.com/apache/kvrocks-controller/store"
)

const (
	TypeDisabled = "disabled"
	TypeLocal    = "local"
)

type Service struct {
	cfg *config.AuthConfig
	s   *store.ClusterStore
}

type UserInfo struct {
	Username  string         `json:"username"`
	Role      store.UserRole `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Principal struct {
	Username string         `json:"username"`
	Role     store.UserRole `json:"role"`
}

type LoginResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type Claims struct {
	Role store.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func NewService(cfg *config.AuthConfig, s *store.ClusterStore) *Service {
	if cfg == nil {
		cfg = config.DefaultAuthConfig()
	}
	return &Service{cfg: cfg, s: s}
}

func (svc *Service) Enabled() bool {
	if svc == nil || svc.cfg == nil {
		return false
	}
	return strings.EqualFold(svc.cfg.Type, TypeLocal)
}

func (svc *Service) Bootstrap(ctx context.Context) error {
	if !svc.Enabled() {
		return nil
	}

	username := strings.TrimSpace(svc.cfg.DefaultAdminUsername)
	if username == "" {
		return fmt.Errorf("auth default admin username: %w", consts.ErrInvalidArgument)
	}

	if _, err := svc.s.GetUser(ctx, username); err == nil {
		return nil
	} else if !errors.Is(err, consts.ErrNotFound) {
		return err
	}

	passwordHash, err := hashPassword(svc.cfg.DefaultAdminPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return svc.s.CreateUser(ctx, &store.User{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         store.UserRoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (svc *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	if !svc.Enabled() {
		return nil, fmt.Errorf("auth is disabled: %w", consts.ErrForbidden)
	}
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required: %w", consts.ErrInvalidArgument)
	}

	user, err := svc.s.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, consts.ErrUnauthorized
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, consts.ErrUnauthorized
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(svc.cfg.JWTTokenTTLSeconds) * time.Second)

	token, err := svc.signToken(user, expiresAt, now)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      toUserInfo(user),
	}, nil
}

func (svc *Service) Authenticate(_ context.Context, tokenString string) (*Principal, error) {
	if !svc.Enabled() {
		return nil, nil
	}
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, consts.ErrUnauthorized
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, consts.ErrUnauthorized
		}
		return []byte(svc.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, consts.ErrUnauthorized
	}
	if claims.Subject == "" {
		return nil, consts.ErrUnauthorized
	}
	if err := claims.Role.Validate(); err != nil {
		return nil, consts.ErrUnauthorized
	}
	return &Principal{
		Username: claims.Subject,
		Role:     claims.Role,
	}, nil
}

func (svc *Service) ListUsers(ctx context.Context) ([]UserInfo, error) {
	users, err := svc.s.ListUser(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]UserInfo, 0, len(users))
	for _, user := range users {
		infos = append(infos, toUserInfo(user))
	}
	return infos, nil
}

func (svc *Service) GetUser(ctx context.Context, username string) (*UserInfo, error) {
	user, err := svc.s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	info := toUserInfo(user)
	return &info, nil
}

func (svc *Service) CreateUser(ctx context.Context, username, password string, role store.UserRole) (*UserInfo, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required: %w", consts.ErrInvalidArgument)
	}
	if err := role.Validate(); err != nil {
		return nil, err
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	user := &store.User{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := svc.s.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	info := toUserInfo(user)
	return &info, nil
}

func (svc *Service) UpdateUser(ctx context.Context, username string, role *store.UserRole, password string) (*UserInfo, error) {
	user, err := svc.s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	if role != nil {
		if err := role.Validate(); err != nil {
			return nil, err
		}
		if user.Role == store.UserRoleAdmin && *role != store.UserRoleAdmin {
			if err := svc.ensureAnotherAdmin(ctx, username); err != nil {
				return nil, err
			}
		}
		user.Role = *role
	}
	if password != "" {
		passwordHash, err := hashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = passwordHash
	}
	user.UpdatedAt = time.Now().UTC()
	if err := svc.s.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	info := toUserInfo(user)
	return &info, nil
}

func (svc *Service) RemoveUser(ctx context.Context, username string) error {
	user, err := svc.s.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if user.Role == store.UserRoleAdmin {
		if err := svc.ensureAnotherAdmin(ctx, username); err != nil {
			return err
		}
	}
	return svc.s.RemoveUser(ctx, username)
}

func (svc *Service) signToken(user *store.User, expiresAt time.Time, now time.Time) (string, error) {
	claims := &Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(svc.cfg.JWTSecret))
}

func (svc *Service) ensureAnotherAdmin(ctx context.Context, username string) error {
	users, err := svc.s.ListUser(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Username != username && user.Role == store.UserRoleAdmin {
			return nil
		}
	}
	return fmt.Errorf("cannot remove the last administrator: %w", consts.ErrForbidden)
}

func toUserInfo(user *store.User) UserInfo {
	return UserInfo{
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func hashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return "", fmt.Errorf("password is required: %w", consts.ErrInvalidArgument)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
