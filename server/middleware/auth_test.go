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

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/apache/kvrocks-controller/auth"
	"github.com/apache/kvrocks-controller/config"
	"github.com/apache/kvrocks-controller/store"
	"github.com/apache/kvrocks-controller/store/engine"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := auth.NewService(&config.AuthConfig{
		Type:                      auth.TypeLocal,
		JWTSecret:                 "test-secret",
		MaxSessionDurationSeconds: 3600,
		DefaultAdminUsername:      "admin",
		DefaultAdminPassword:      "admin-password",
	}, store.NewClusterStore(engine.NewMock()))
	require.NoError(t, svc.Bootstrap(context.Background()))
	loginResult, err := svc.Login(context.Background(), "admin", "admin-password")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/protected", AuthRequired(svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+loginResult.Token)
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestAdminRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := auth.NewService(&config.AuthConfig{
		Type:                      auth.TypeLocal,
		JWTSecret:                 "test-secret",
		MaxSessionDurationSeconds: 3600,
		DefaultAdminUsername:      "admin",
		DefaultAdminPassword:      "admin-password",
	}, store.NewClusterStore(engine.NewMock()))
	require.NoError(t, svc.Bootstrap(context.Background()))
	_, err := svc.CreateUser(context.Background(), "dev", "dev-password", store.UserRoleUser)
	require.NoError(t, err)
	loginResult, err := svc.Login(context.Background(), "dev", "dev-password")
	require.NoError(t, err)

	router := gin.New()
	router.GET("/admin", AuthRequired(svc), AdminRequired(svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+loginResult.Token)
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAuthEnabledRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := auth.NewService(config.DefaultAuthConfig(), store.NewClusterStore(engine.NewMock()))

	router := gin.New()
	router.GET("/user", AuthEnabledRequired(svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusForbidden, recorder.Code)
}
