package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"

	"ssh-pro/storage"
)

type hostItem struct {
	host storage.Host
}

func (h hostItem) Title() string { return h.host.Name }

func (h hostItem) Description() string {
	return fmt.Sprintf("%s • %s:%d", h.host.User, h.host.IP, h.host.PortOrDefault())
}

func (h hostItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s", h.host.Name, h.host.User, h.host.IP)
}

func listItemsFromHosts(hosts []storage.Host) []list.Item {
	items := make([]list.Item, 0, len(hosts))
	for _, host := range hosts {
		items = append(items, hostItem{host: host})
	}
	return items
}
