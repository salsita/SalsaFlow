/*
   Copyright (C) 2014  Salsita s.r.o.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program. If not, see {http://www.gnu.org/licenses/}.
*/

package client

type User struct {
	Self         string `json:"self,omitempty"`
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	AvatarURLs   struct {
		Size16 string `json:"16x16,omitempty"`
		Size24 string `json:"24x24,omitempty"`
		Size32 string `json:"32x32,omitempty"`
		Size48 string `json:"48x48,omitempty"`
	} `json:"avatarUrls,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active,omitempty"`
	TimeZone    string `json:"timeZone,omitempty"`
	Groups      struct {
		Size  int `json:"size,omitempty"`
		Items []struct {
			Name string `json:"name,omitempty"`
			Self string `json:"self,omitempty"`
		} `json:"items,omitempty"`
	} `json:"groups,omitempty"`
}
