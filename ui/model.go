package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssh-pro/config"
	sshcmd "ssh-pro/ssh"
	"ssh-pro/storage"
)

type mode int

const (
	modeList mode = iota
	modeForm
	modeTheme
)

type sshFinishedMsg struct {
	err error
}

// Model is the Bubble Tea model for the SSH host manager.
type Model struct {
	store *storage.Store
	hosts []storage.Host
	list      list.Model
	themeList list.Model
	mode      mode

	themeConfig config.ThemeConfig

	inputs    []textinput.Model
	focus     int
	editIndex int
	formError     string
	status        string
	confirmDelete bool
	confirmIndex  int
	confirmName   string

	styles styles
	width  int
	height int
}

// NewModel creates a new UI model.
func NewModel(store *storage.Store, hosts []storage.Host, themeConfig config.ThemeConfig) Model {
	items := listItemsFromHosts(hosts)

	theme, ok := config.FindTheme(themeConfig, themeConfig.Current)
	if !ok {
		theme = config.DefaultTheme()
		themeConfig.Current = theme.Name
	}

	delegate := listDelegateFromTheme(theme)
	l := list.New(items, delegate, 0, 0)
	l.Title = "Hosts disponibles:"
	l.SetShowTitle(true)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.FilterInput.Prompt = "Buscar: "
	l.FilterInput.Placeholder = "escribe para filtrar"
	l.SetShowPagination(true)
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterPrompt))
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterText))

	themeItems := themeItemsFromConfig(themeConfig)
	themeList := list.New(themeItems, listDelegateFromTheme(theme), 0, 0)
	themeList.Title = "Temas disponibles:"
	themeList.SetShowTitle(true)
	themeList.SetShowHelp(false)
	themeList.SetShowStatusBar(false)
	themeList.SetFilteringEnabled(true)
	themeList.SetShowFilter(true)
	themeList.FilterInput.Prompt = "Buscar: "
	themeList.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterPrompt))
	themeList.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterText))
	themeList.FilterInput.Placeholder = "escribe para filtrar"
	themeList.SetShowPagination(true)

	model := Model{
		store:        store,
		hosts:        hosts,
		list:         l,
		themeList:    themeList,
		mode:         modeList,
		themeConfig:  themeConfig,
		editIndex:    -1,
		confirmIndex: -1,
		styles:       stylesFromTheme(theme),
	}

	return model
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeList:
		return m.updateList(msg)
	case modeForm:
		return m.updateForm(msg)
	case modeTheme:
		return m.updateTheme(msg)
	default:
		return m, nil
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.mode == modeForm {
		return m.viewForm()
	}
	if m.mode == modeTheme {
		return m.viewTheme()
	}
	return m.viewList()
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateListSize()
		m.updateThemeListSize()
	case sshFinishedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("SSH terminó con error: %v", msg.err)
		} else {
			m.status = "Sesión SSH finalizada."
		}
		return m, nil
	case tea.KeyMsg:
		if m.list.SettingFilter() {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			break
		}
		if m.confirmDelete {
			switch msg.String() {
			case "y", "enter":
				if m.confirmIndex >= 0 {
					if err := m.deleteHost(m.confirmIndex); err != nil {
						m.status = fmt.Sprintf("Error al eliminar: %v", err)
					}
				}
				m.confirmDelete = false
				m.confirmIndex = -1
				m.confirmName = ""
				return m, nil
			case "n", "esc":
				m.confirmDelete = false
				m.confirmIndex = -1
				m.confirmName = ""
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.startForm(storage.Host{}, -1)
			return m, nil
		case "e":
			if host, idx, ok := m.selectedHost(); ok {
				m.startForm(host, idx)
			}
			return m, nil
		case "d":
			if host, idx, ok := m.selectedHost(); ok {
				m.confirmDelete = true
				m.confirmIndex = idx
				m.confirmName = host.Name
				m.status = ""
			}
			return m, nil
		case "t":
			m.startThemePicker()
			return m, nil
		case "enter":
			if host, _, ok := m.selectedHost(); ok {
				cmd, err := sshcmd.BuildCommand(host)
				if err != nil {
					m.status = fmt.Sprintf("Error al preparar SSH: %v", err)
					return m, nil
				}
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return sshFinishedMsg{err: err}
				})
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = modeList
			m.formError = ""
			return m, nil
		case "tab", "down":
			m.setFocus(m.focus + 1)
			return m, nil
		case "shift+tab", "up":
			m.setFocus(m.focus - 1)
			return m, nil
		case "enter":
			if m.focus < len(m.inputs)-1 {
				m.setFocus(m.focus + 1)
				return m, nil
			}
			host, err := m.hostFromForm()
			if err != nil {
				m.formError = err.Error()
				return m, nil
			}
			if err := m.saveHost(host); err != nil {
				m.formError = err.Error()
				return m, nil
			}
			m.mode = modeList
			m.formError = ""
			return m, nil
		}
	}

	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateTheme(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateThemeListSize()
	case tea.KeyMsg:
		if m.themeList.SettingFilter() {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			break
		}
		switch msg.String() {
		case "esc":
			m.mode = modeList
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if item, ok := m.themeList.SelectedItem().(themeItem); ok {
				if err := m.applyTheme(item.theme); err != nil {
					m.status = fmt.Sprintf("No se pudo aplicar el tema: %v", err)
				} else {
					m.status = fmt.Sprintf("Tema %q aplicado.", item.theme.Name)
				}
			}
			m.mode = modeList
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.themeList, cmd = m.themeList.Update(msg)
	return m, cmd
}

func (m *Model) updateListSize() {
	contentWidth, contentHeight := m.contentSize()
	width := max(20, contentWidth)
	height := max(6, contentHeight-listReservedLines())
	m.list.SetSize(width, height)
}

func (m *Model) updateThemeListSize() {
	contentWidth, contentHeight := m.contentSize()
	width := max(20, contentWidth)
	height := max(6, contentHeight-listReservedLines())
	m.themeList.SetSize(width, height)
}

func (m *Model) startThemePicker() {
	m.mode = modeTheme
	m.themeList.SetItems(themeItemsFromConfig(m.themeConfig))
	m.updateThemeListSize()
	for i, theme := range m.themeConfig.Themes {
		if theme.Name == m.themeConfig.Current {
			m.themeList.Select(i)
			break
		}
	}
}

func (m *Model) applyTheme(theme config.Theme) error {
	m.themeConfig.Current = theme.Name
	if err := config.SaveThemeConfig(m.themeConfig); err != nil {
		return err
	}

	m.styles = stylesFromTheme(theme)

	delegate := listDelegateFromTheme(theme)
	m.list.SetDelegate(delegate)
	m.themeList.SetDelegate(delegate)

	m.list.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterPrompt))
	m.list.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterText))
	m.themeList.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterPrompt))
	m.themeList.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.FilterText))

	m.themeList.SetItems(themeItemsFromConfig(m.themeConfig))

	if m.mode == modeForm {
		m.setFocus(m.focus)
	}

	return nil
}

func (m *Model) startForm(host storage.Host, editIndex int) {
	m.editIndex = editIndex
	m.formError = ""
	m.status = ""

	m.inputs = []textinput.Model{
		newTextInput("Nombre: ", "Servidor-Produccion"),
		newTextInput("Usuario: ", "ubuntu"),
		newTextInput("Host/IP: ", "192.168.1.50"),
		newTextInput("Llave SSH (opcional): ", "id_ed25519"),
		newTextInput("Puerto: ", "22"),
	}

	m.inputs[0].SetValue(host.Name)
	m.inputs[1].SetValue(host.User)
	m.inputs[2].SetValue(host.IP)
	m.inputs[3].SetValue(host.Key)
	if editIndex >= 0 {
		m.inputs[4].SetValue(strconv.Itoa(host.PortOrDefault()))
	}

	m.focus = 0
	m.setFocus(m.focus)
	m.mode = modeForm
}

func (m *Model) setFocus(index int) {
	if len(m.inputs) == 0 {
		return
	}
	if index < 0 {
		index = len(m.inputs) - 1
	}
	if index >= len(m.inputs) {
		index = 0
	}
	m.focus = index
	for i := range m.inputs {
		if i == m.focus {
			m.inputs[i].PromptStyle = m.styles.focusedInput
			m.inputs[i].TextStyle = m.styles.focusedInput
			m.inputs[i].Focus()
		} else {
			m.inputs[i].PromptStyle = m.styles.blurredInput
			m.inputs[i].TextStyle = m.styles.blurredInput
			m.inputs[i].Blur()
		}
	}
}

func (m *Model) hostFromForm() (storage.Host, error) {
	name := strings.TrimSpace(m.inputs[0].Value())
	user := strings.TrimSpace(m.inputs[1].Value())
	ip := strings.TrimSpace(m.inputs[2].Value())
	key := strings.TrimSpace(m.inputs[3].Value())
	portValue := strings.TrimSpace(m.inputs[4].Value())

	if name == "" || user == "" || ip == "" {
		return storage.Host{}, errors.New("nombre, usuario y host/IP son obligatorios")
	}

	for i, host := range m.hosts {
		if i == m.editIndex {
			continue
		}
		if host.Name == name {
			return storage.Host{}, fmt.Errorf("ya existe un host con el nombre %q", name)
		}
	}

	port := 22
	if portValue != "" {
		parsed, err := strconv.Atoi(portValue)
		if err != nil || parsed < 1 || parsed > 65535 {
			return storage.Host{}, errors.New("el puerto debe ser un número entre 1 y 65535")
		}
		port = parsed
	}

	return storage.Host{
		Name: name,
		User: user,
		IP:   ip,
		Key:  key,
		Port: port,
	}, nil
}

func (m *Model) saveHost(host storage.Host) error {
	if m.editIndex >= 0 {
		m.hosts[m.editIndex] = host
	} else {
		m.hosts = append(m.hosts, host)
		m.editIndex = len(m.hosts) - 1
	}

	if err := m.store.Save(m.hosts); err != nil {
		return err
	}

	m.list.SetItems(listItemsFromHosts(m.hosts))
	if m.editIndex >= 0 && m.editIndex < len(m.hosts) {
		m.list.Select(m.editIndex)
	}
	m.status = "Host guardado."
	return nil
}

func (m *Model) deleteHost(index int) error {
	if index < 0 || index >= len(m.hosts) {
		return nil
	}
	name := m.hosts[index].Name
	m.hosts = append(m.hosts[:index], m.hosts[index+1:]...)

	if err := m.store.Save(m.hosts); err != nil {
		return err
	}

	m.list.SetItems(listItemsFromHosts(m.hosts))
	if len(m.hosts) > 0 {
		if index >= len(m.hosts) {
			index = len(m.hosts) - 1
		}
		m.list.Select(index)
	}
	m.status = fmt.Sprintf("Host %q eliminado.", name)
	return nil
}

func (m *Model) selectedHost() (storage.Host, int, bool) {
	if len(m.hosts) == 0 {
		return storage.Host{}, -1, false
	}
	idx := m.list.Index()
	if idx < 0 || idx >= len(m.hosts) {
		return storage.Host{}, -1, false
	}
	return m.hosts[idx], idx, true
}

func newTextInput(prompt, placeholder string) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Prompt = prompt
	input.CharLimit = 255
	input.Width = 50
	return input
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
