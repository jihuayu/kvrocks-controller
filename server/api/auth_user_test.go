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

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/apache/kvrocks-controller/auth"
	"github.com/apache/kvrocks-controller/config"
	"github.com/apache/kvrocks-controller/consts"
	"github.com/apache/kvrocks-controller/store"
	"github.com/apache/kvrocks-controller/store/engine"
)

func newTestAuthService(t *testing.T) *auth.Service {
	t.Helper()
	svc := auth.NewService(&config.AuthConfig{
		Type:                      auth.TypeLocal,
		JWTSecret:                 "test-secret",
		MaxSessionDurationSeconds: 3600,
		DefaultAdminUsername:      "admin",
		DefaultAdminPassword:      "admin-password",
	}, store.NewClusterStore(engine.NewMock()))
	require.NoError(t, svc.Bootstrap(context.Background()))
	return svc
}

func TestAuthHandlerLoginMeAndLogout(t *testing.T) {
	svc := newTestAuthService(t)
	handler := &AuthHandler{auth: svc}

	recorder := httptest.NewRecorder()
	ctx := GetTestContext(recorder)
	ctx.Request.Body = io.NopCloser(bytes.NewBufferString(`{"username":"admin","password":"admin-password"}`))
	handler.Login(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)

	var loginRsp struct {
		Data auth.LoginResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &loginRsp))
	require.NotEmpty(t, loginRsp.Data.Token)

	user, session, err := svc.Authenticate(context.Background(), loginRsp.Data.Token)
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	ctx = GetTestContext(recorder)
	ctx.Set(consts.ContextKeyAuthUser, user)
	ctx.Set(consts.ContextKeyAuthSession, session)
	handler.Me(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	ctx = GetTestContext(recorder)
	ctx.Set(consts.ContextKeyAuthSession, session)
	handler.Logout(ctx)
	require.Equal(t, http.StatusNoContent, recorder.Code)

	_, _, err = svc.Authenticate(context.Background(), loginRsp.Data.Token)
	require.ErrorIs(t, err, consts.ErrUnauthorized)
}

func TestUserHandlerBasics(t *testing.T) {
	svc := newTestAuthService(t)
	handler := &UserHandler{auth: svc}

	recorder := httptest.NewRecorder()
	ctx := GetTestContext(recorder)
	ctx.Request.Body = io.NopCloser(bytes.NewBufferString(`{"username":"dev","password":"dev-password","role":"user"}`))
	handler.Create(ctx)
	require.Equal(t, http.StatusCreated, recorder.Code)

	recorder = httptest.NewRecorder()
	ctx = GetTestContext(recorder)
	handler.List(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "password_hash")

	role := store.UserRoleAdmin
	body, err := json.Marshal(gin.H{"role": role})
	require.NoError(t, err)
	recorder = httptest.NewRecorder()
	ctx = GetTestContext(recorder)
	ctx.Params = []gin.Param{{Key: "username", Value: "dev"}}
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	handler.Update(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	ctx = GetTestContext(recorder)
	ctx.Params = []gin.Param{{Key: "username", Value: "dev"}}
	handler.Remove(ctx)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}
