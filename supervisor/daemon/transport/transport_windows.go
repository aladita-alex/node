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

package transport

import (
	"fmt"
	"log"
	"strings"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows/svc"
)

const sock = `\\.\pipe\mystpipe`

// Start starts a listener on a unix domain socket.
// Conversation is handled by the handlerFunc.
func Start(handle handlerFunc) error {
	return svc.Run("WireGuardManager", &managerService{handle: handle})
}

type managerService struct {
	handle handlerFunc
}

func (m *managerService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	go func() {
		if err := m.listenPipe(); err != nil {
			log.Printf("could not listen pipe: %v", err)
		}
	}()

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				return
			case svc.Pause:
				s <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				log.Printf("unexpected control request #%d", c)
			}
		}
	}
}

func (m *managerService) listenPipe() error {
	// Allow Administrators and SYSTEM, plus whatever additional users or groups are specified.
	socketGroup := "Users"
	sddl := "D:P(A;;GA;;;BA)(A;;GA;;;SY)"
	for _, g := range strings.Split(socketGroup, ",") {
		sid, err := winio.LookupSidByName(g)
		if err != nil {
			return err
		}
		sddl += fmt.Sprintf("(A;;GRGW;;;%s)", sid)
	}
	c := winio.PipeConfig{
		SecurityDescriptor: sddl,
		MessageMode:        true,  // Use message mode so that CloseWrite() is supported.
		InputBufferSize:    65536, // Use 64KB buffers to improve performance.
		OutputBufferSize:   65536,
	}

	l, err := winio.ListenPipe(sock, &c)
	if err != nil {
		return fmt.Errorf("error listening: %w", err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Println("Error closing listener:", err)
		}
	}()
	for {
		log.Println("Waiting for connections...")
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accept error: %w", err)
		}
		go func() {
			peer := conn.RemoteAddr().Network()
			log.Println("Client connected:", peer)
			m.handle(conn)
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection for:", peer, err)
			}
			log.Println("Client disconnected:", peer)
		}()
	}
}
