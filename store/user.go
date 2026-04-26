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

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/apache/kvrocks-controller/consts"
)

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (r UserRole) Validate() error {
	switch r {
	case UserRoleAdmin, UserRoleUser:
		return nil
	default:
		return fmt.Errorf("unknown user role %q: %w", r, consts.ErrInvalidArgument)
	}
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("username is required: %w", consts.ErrInvalidArgument)
	}
	if strings.Contains(u.Username, "/") {
		return fmt.Errorf("username must not contain '/': %w", consts.ErrInvalidArgument)
	}
	if strings.TrimSpace(u.PasswordHash) == "" {
		return fmt.Errorf("password hash is required: %w", consts.ErrInvalidArgument)
	}
	if err := u.Role.Validate(); err != nil {
		return err
	}
	return nil
}

func (s *Session) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("session id is required: %w", consts.ErrInvalidArgument)
	}
	if strings.Contains(s.ID, "/") {
		return fmt.Errorf("session id must not contain '/': %w", consts.ErrInvalidArgument)
	}
	if strings.TrimSpace(s.Username) == "" {
		return fmt.Errorf("session username is required: %w", consts.ErrInvalidArgument)
	}
	if s.ExpiresAt.IsZero() {
		return fmt.Errorf("session expires_at is required: %w", consts.ErrInvalidArgument)
	}
	return nil
}

func (s *ClusterStore) ListUser(ctx context.Context) ([]*User, error) {
	entries, err := s.e.List(ctx, userPrefix)
	if err != nil {
		return nil, err
	}
	users := make([]*User, 0, len(entries))
	for _, entry := range entries {
		user, err := s.GetUser(ctx, entry.Key)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *ClusterStore) GetUser(ctx context.Context, username string) (*User, error) {
	value, err := s.e.Get(ctx, buildUserKey(username))
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	var user User
	if err := json.Unmarshal(value, &user); err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}
	return &user, nil
}

func (s *ClusterStore) CreateUser(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if exists, _ := s.e.Exists(ctx, buildUserKey(user.Username)); exists {
		return fmt.Errorf("user: %w", consts.ErrAlreadyExists)
	}
	return s.setUser(ctx, user)
}

func (s *ClusterStore) UpdateUser(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if exists, _ := s.e.Exists(ctx, buildUserKey(user.Username)); !exists {
		return fmt.Errorf("user: %w", consts.ErrNotFound)
	}
	return s.setUser(ctx, user)
}

func (s *ClusterStore) RemoveUser(ctx context.Context, username string) error {
	if exists, _ := s.e.Exists(ctx, buildUserKey(username)); !exists {
		return fmt.Errorf("user: %w", consts.ErrNotFound)
	}
	return s.e.Delete(ctx, buildUserKey(username))
}

func (s *ClusterStore) CreateSession(ctx context.Context, session *Session) error {
	if err := session.Validate(); err != nil {
		return err
	}
	if exists, _ := s.e.Exists(ctx, buildSessionKey(session.ID)); exists {
		return fmt.Errorf("session: %w", consts.ErrAlreadyExists)
	}
	value, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}
	return s.e.Set(ctx, buildSessionKey(session.ID), value)
}

func (s *ClusterStore) GetSession(ctx context.Context, id string) (*Session, error) {
	value, err := s.e.Get(ctx, buildSessionKey(id))
	if err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}
	var session Session
	if err := json.Unmarshal(value, &session); err != nil {
		return nil, fmt.Errorf("session: %w", err)
	}
	return &session, nil
}

func (s *ClusterStore) RemoveSession(ctx context.Context, id string) error {
	return s.e.Delete(ctx, buildSessionKey(id))
}

func (s *ClusterStore) setUser(ctx context.Context, user *User) error {
	value, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("user: %w", err)
	}
	return s.e.Set(ctx, buildUserKey(user.Username), value)
}
