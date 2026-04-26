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
	"github.com/gin-gonic/gin"

	"github.com/apache/kvrocks-controller/auth"
	"github.com/apache/kvrocks-controller/consts"
	"github.com/apache/kvrocks-controller/server/helper"
	"github.com/apache/kvrocks-controller/store"
)

type AuthHandler struct {
	auth *auth.Service
}

func (handler *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ResponseBadRequest(c, err)
		return
	}

	result, err := handler.auth.Login(c, req.Username, req.Password)
	if err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseOK(c, result)
}

func (handler *AuthHandler) Logout(c *gin.Context) {
	if handler.auth == nil || !handler.auth.Enabled() {
		helper.ResponseNoContent(c)
		return
	}
	session, _ := c.MustGet(consts.ContextKeyAuthSession).(*store.Session)
	if err := handler.auth.Logout(c, session.ID); err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseNoContent(c)
}

func (handler *AuthHandler) Me(c *gin.Context) {
	if handler.auth == nil || !handler.auth.Enabled() {
		helper.ResponseOK(c, gin.H{"auth_enabled": false})
		return
	}
	principal, _ := c.MustGet(consts.ContextKeyAuthUser).(*auth.Principal)
	helper.ResponseOK(c, gin.H{
		"auth_enabled": true,
		"user":         principal,
	})
}
