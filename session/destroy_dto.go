/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package session

// DestroyRequest structure represents message from service consumer to destroy session for given session id
type DestroyRequest struct {
	SessionID string `json:"session_id"`
}

// DestroyResponse structure represents service provider response to given session request from consumer
type DestroyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
