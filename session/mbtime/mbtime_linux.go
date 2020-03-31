/*
 * Copyright (C) 2020 The "MysteriumNetwork/node" Authors.
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

package mbtime

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

func nanotime() (uint64, error) {
	var ts unix.Timespec
	log.Info().Msgf("Getting nanotime")
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &ts); err != nil {
		if tempSyscallErr(err) {
			if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &ts); err != nil {
				return 0, err
			}
		}
	}

	log.Info().Msgf("nanotimeRes ts.Sec*1e9=%v, ts.Nsec=", ts.Sec*1e9, ts.Nsec)
	return uint64(ts.Sec*1e9 + ts.Nsec), nil
}
