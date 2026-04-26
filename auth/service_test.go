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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/apache/kvrocks-controller/config"
	"github.com/apache/kvrocks-controller/consts"
	"github.com/apache/kvrocks-controller/store"
	"github.com/apache/kvrocks-controller/store/engine"
)

func newTestService() *Service {
	return NewService(&config.AuthConfig{
		Type:                 TypeLocal,
		JWTSecret:            "test-secret",
		JWTTokenTTLSeconds:   3600,
		DefaultAdminUsername: "admin",
		DefaultAdminPassword: "admin-password",
	}, store.NewClusterStore(engine.NewMock()))
}

func TestServiceLoginAndAuthenticate(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	require.NoError(t, svc.Bootstrap(ctx))

	result, err := svc.Login(ctx, "admin", "admin-password")
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, "admin", result.User.Username)
	require.Equal(t, store.UserRoleAdmin, result.User.Role)

	principal, err := svc.Authenticate(ctx, result.Token)
	require.NoError(t, err)
	require.Equal(t, "admin", principal.Username)
	require.Equal(t, store.UserRoleAdmin, principal.Role)
}

func TestServiceRejectsInvalidPassword(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	require.NoError(t, svc.Bootstrap(ctx))

	_, err := svc.Login(ctx, "admin", "wrong-password")
	require.ErrorIs(t, err, consts.ErrUnauthorized)
}

func TestServiceProtectsLastAdmin(t *testing.T) {
	ctx := context.Background()
	svc := newTestService()
	require.NoError(t, svc.Bootstrap(ctx))

	err := svc.RemoveUser(ctx, "admin")
	require.True(t, errors.Is(err, consts.ErrForbidden))

	role := store.UserRoleUser
	_, err = svc.UpdateUser(ctx, "admin", &role, "")
	require.True(t, errors.Is(err, consts.ErrForbidden))

	_, err = svc.CreateUser(ctx, "second-admin", "password", store.UserRoleAdmin)
	require.NoError(t, err)
	_, err = svc.UpdateUser(ctx, "admin", &role, "")
	require.NoError(t, err)
}
