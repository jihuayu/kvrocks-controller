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
	"github.com/apache/kvrocks-controller/server/helper"
	"github.com/apache/kvrocks-controller/store"
)

type UserHandler struct {
	auth *auth.Service
}

func (handler *UserHandler) List(c *gin.Context) {
	users, err := handler.auth.ListUsers(c)
	if err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseOK(c, gin.H{"users": users})
}

func (handler *UserHandler) Get(c *gin.Context) {
	user, err := handler.auth.GetUser(c, c.Param("username"))
	if err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseOK(c, gin.H{"user": user})
}

func (handler *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string         `json:"username" binding:"required"`
		Password string         `json:"password" binding:"required"`
		Role     store.UserRole `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ResponseBadRequest(c, err)
		return
	}
	user, err := handler.auth.CreateUser(c, req.Username, req.Password, req.Role)
	if err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseCreated(c, gin.H{"user": user})
}

func (handler *UserHandler) Update(c *gin.Context) {
	var req struct {
		Password string          `json:"password"`
		Role     *store.UserRole `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ResponseBadRequest(c, err)
		return
	}
	user, err := handler.auth.UpdateUser(c, c.Param("username"), req.Role, req.Password)
	if err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseOK(c, gin.H{"user": user})
}

func (handler *UserHandler) Remove(c *gin.Context) {
	if err := handler.auth.RemoveUser(c, c.Param("username")); err != nil {
		helper.ResponseError(c, err)
		return
	}
	helper.ResponseNoContent(c)
}
