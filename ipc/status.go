// This file is part of autosr.
//
// autosr is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// autosr is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with autosr.  If not, see <https://www.gnu.org/licenses/>.

package ipc

import (
	"github.com/bobbytrapz/autosr/track"
)

// Dashboard represents a connected dashboard
type Dashboard struct {
	SelectURL string
	Tracking  []track.Info
}

var status Dashboard

func replicate(req *Dashboard, res *Dashboard) {
	if req.SelectURL == "?" {
		res.SelectURL = status.SelectURL
	} else {
		status.SelectURL = req.SelectURL
	}

	lst := track.ListTracking()
	res.Tracking = make([]track.Info, len(lst))
	copy(res.Tracking, lst)
}

// Status for the dashboard
func (c *Command) Status(req *Dashboard, res *Dashboard) error {
	replicate(req, res)

	return nil
}
